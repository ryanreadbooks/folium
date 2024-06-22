package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ryanreadbooks/folium/internal/pkg"
)

var (
	ErrNilAlloc = pkg.ErrInvalidArgs.Message("alloc arg is nil")
)

func QueryByKey(ctx context.Context, key string) (*Alloc, error) {
	var alloc Alloc
	query := fmt.Sprintf(
		`select %s from %s where biz_key = ?`,
		allocColumns,
		TableName,
	)
	row := db.QueryRowContext(ctx, query, key)
	err := row.Scan(&alloc.Id,
		&alloc.Key,
		&alloc.CurId,
		&alloc.Step,
		&alloc.CreatedAt,
		&alloc.UpdatedAt)
	if err != nil {
		log.Printf("dao query row err: %v\n", err)
		return nil, pkg.ErrDb.Message(err.Error())
	}

	return &alloc, nil
}

// QueryAll retrieves all the alloc records from db
func QueryAll(ctx context.Context) ([]*Alloc, error) {
	query := fmt.Sprintf(
		`select %s from %s`, allocColumns, TableName,
	)

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("dao query all rows err: %v\n", err)
		return nil, pkg.ErrDb.Message(err.Error())
	}
	defer rows.Close()

	var allocs []*Alloc
	for rows.Next() {
		var alloc Alloc
		err := rows.Scan(&alloc.Id,
			&alloc.Key,
			&alloc.CurId,
			&alloc.Step,
			&alloc.CreatedAt,
			&alloc.UpdatedAt)
		if err != nil {
			log.Printf("dao query all rows scan err: %v\n", err)
			return nil, pkg.ErrDb.Message(err.Error())
		}
		allocs = append(allocs, &alloc)
	}

	if err := rows.Err(); err != nil {
		return nil, pkg.ErrDb.Message(err.Error())
	}

	return allocs, nil
}

func QueryAllKeys(ctx context.Context) ([]string, error) {
	query := fmt.Sprintf("select biz_key from %s", TableName)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("dao query all keys err: %v\n", err)
		return nil, pkg.ErrDb.Message(err.Error())
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		err := rows.Scan(&key)
		if err != nil {
			log.Printf("dao query all keys scan err: %v\n", err)
			return nil, err
		}
		keys = append(keys, key)
	}

	if err := rows.Err(); err != nil {
		return nil, pkg.ErrDb.Message(err.Error())
	}

	return keys, nil
}

// create or update the alloc given the specific key
func CreateUpdate(ctx context.Context, alloc *Alloc) error {
	if alloc == nil {
		return ErrNilAlloc
	}

	if alloc.Step == 0 {
		alloc.Step = defaultStep
	}

	statement := `
		insert into %s(biz_key, cur_id, step, created_at, updated_at)
		values (?,?,?,?,?) as new_vals
		on duplicate key update
		cur_id = %s.cur_id + new_vals.step,
		step = new_vals.step,
		updated_at = ?
	`

	statement = fmt.Sprintf(statement, TableName, TableName)
	now := time.Now().UnixMilli()
	return stmtExec(ctx, statement, alloc.Key, alloc.CurId, alloc.Step, now, now, now)
}

type TakeIdResult struct {
	Begin uint64
	End   uint64
	Step  uint32
}

// return curId before update
// query alloc with specific key, then update the corresponding records
// [Begin, End) is allowed
func TakeIdForKey(ctx context.Context, key string, newStep uint32) (*TakeIdResult, error) {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("dao begin tx err: %v\n", err)
		return nil, pkg.ErrDb.Message(err.Error())
	}

	var (
		rollback = true
	)

	defer func() {
		if rollback {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	row, err := tx.QueryContext(
		ctx,
		fmt.Sprintf("select cur_id, step from %s where biz_key = ? limit 1 for update", TableName),
		key,
	)

	if newStep == 0 {
		newStep = defaultStep
	}

	var (
		// if key is not found in db, the following will be the default retvals
		curId uint64 = defaultCurId
		step  uint32 = newStep
	)

	if err != nil {
		if !errors.Is(sql.ErrNoRows, err) {
			log.Printf("dao tx query err: %v\n", err)
			return nil, pkg.ErrDb.Message(err.Error())
		}
	} else {
		for row.Next() {
			err = row.Scan(&curId, &step)
			if err != nil {
				log.Printf("dao scan row err: %v\n", err)
				return nil, pkg.ErrDb.Message(err.Error())
			}
			break
		}

		if err := row.Err(); err != nil {
			log.Printf("dao row err: %v\n", err)
			return nil, pkg.ErrDb.Message(err.Error())
		}
		row.Close() // close row explicitly
	}

	// initialization and updating
	statement := `
		insert into %s(biz_key, cur_id, step, created_at, updated_at)
		values (?,?,?,?,?) as new_vals
		on duplicate key update
		cur_id = %s.cur_id + new_vals.step,
		step = new_vals.step,
		updated_at = ?
	`

	statement = fmt.Sprintf(statement, TableName, TableName)

	now := time.Now().UnixMilli()
	// update cur_id = cur_id + step
	err = txStmtExec(ctx, tx,
		statement,
		key, defaultCurId+uint64(newStep), newStep, now, now, now,
	)
	if err != nil {
		log.Printf("dao tx stmt exec err: %v\n", err)
		return nil, pkg.ErrDb.Message(err.Error())
	}

	rollback = false
	return &TakeIdResult{
		Begin: curId,
		End:   curId + uint64(step),
		Step:  step,
	}, nil
}

func stmtExec(ctx context.Context, statement string, args ...interface{}) error {
	stmt, err := db.PrepareContext(ctx, statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, args...)
	if err != nil {
		return err
	}

	return nil
}

func txStmtExec(ctx context.Context, tx *sql.Tx, statement string, args ...interface{}) error {
	stmt, err := tx.PrepareContext(ctx, statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, args...)
	if err != nil {
		return err
	}

	return nil
}
