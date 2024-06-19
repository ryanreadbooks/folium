package dao

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ryanreadbooks/folium/internal/pkg"

	_ "github.com/go-sql-driver/mysql"
)

const (
	ENV_DB_USER = "ENV_DB_USER"
	ENV_DB_PASS = "ENV_DB_PASS"
	ENV_DB_NAME = "ENV_DB_NAME"
	ENV_DB_ADDR = "ENV_DB_ADDR"
)

var (
	db *sql.DB
)

// init db by environment variables
func InitDB() {
	dsn := getDbDsn(
		os.Getenv(ENV_DB_USER),
		os.Getenv(ENV_DB_PASS),
		os.Getenv(ENV_DB_ADDR),
		os.Getenv(ENV_DB_NAME))

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	log.Println("db inited")
}

func CloseDB() {
	if db != nil {
		err := db.Close()
		if err != nil {
			log.Printf("can not close db: %v", err)
		}
	}
	log.Println("db closed")
}

func GetDB() *sql.DB {
	return db
}

func getDbDsn(user, pass, addr, dbName string) string {
	// [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, addr, dbName)
}

// segment table
const (
	TableName = "alloc_table"
)

type Alloc struct {
	Id        int64  // id primary key
	Key       string // biz_key unique key
	CurId     uint64 // cur_id
	Step      uint32 // step
	Desc      string // description
	CreatedAt int64  // created_at
	UpdatedAt int64  // updated_at
}

var (
	ErrNilAlloc = pkg.ErrInvalidArgs.Message("alloc arg is nil")
)

const (
	allocColumns = "id, biz_key, cur_id, step, description, created_at, updated_at"
)

func QueryByKey(ctx context.Context, key string) (*Alloc, error) {
	var alloc Alloc
	query := fmt.Sprintf(
		`select %s from %s where biz_key = ?`,
		allocColumns,
		TableName,
	)
	row := db.QueryRowContext(ctx, query, key)
	err := row.Scan(&alloc.Id, &alloc.Key, &alloc.CurId, &alloc.Step, &alloc.Desc, &alloc.CreatedAt, &alloc.UpdatedAt)
	if err != nil {
		log.Printf("dao query row err: %v\n", err)
		return nil, pkg.ErrDb.Message(err.Error())
	}

	return &alloc, nil
}

// query alloc with specific key, then update the corresponding records
func ConsumeKey(ctx context.Context, key string) (uint64, uint32, error) {
	return UpdateCurIdWithStepByKey(ctx, key)
}

// QueryAll retrieves all the alloc records from db
func QueryAll(ctx context.Context) ([]*Alloc, error) {
	query := fmt.Sprintf(
		`select %s from %s`,
		allocColumns,
		TableName,
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
		err := rows.Scan(&alloc.Id, &alloc.Key, &alloc.CurId, &alloc.Step, &alloc.Desc, &alloc.CreatedAt, &alloc.UpdatedAt)
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

func Create(ctx context.Context, alloc *Alloc) (int64, error) {
	if alloc == nil {
		return 0, ErrNilAlloc
	}

	query := fmt.Sprintf(
		`insert into %s(
			biz_key,
			cur_id,
			step,
			description,
			created_at,
			updated_at) 
		values (?,?,?,?,?,?)`, TableName,
	)
	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("dao prepare stmt err: %v\n", err)
		return 0, pkg.ErrDb.Message(err.Error())
	}
	defer stmt.Close()

	ctime := time.Now().UnixMilli()
	res, err := stmt.ExecContext(ctx, alloc.Key, alloc.CurId, alloc.Step, alloc.Desc, ctime, ctime)
	if err != nil {
		log.Printf("dao stmt exec err: %v", err)
		return 0, pkg.ErrDb.Message(err.Error())
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		log.Printf("get last inserted id err: %v", err)
		return 0, pkg.ErrDb.Message(err.Error())
	}

	return lastId, nil
}

func UpdateById(ctx context.Context, id int64, alloc *Alloc) error {
	if alloc == nil || id == 0 {
		return pkg.ErrInvalidArgs.Message("invalid args when update")
	}

	statement := fmt.Sprintf(
		"update %s set cur_id = ?, step = ?, description = ?, updated_at = ? where id = ?", TableName,
	)

	return stmtExec(ctx, statement, alloc.CurId, alloc.Step, alloc.Step, time.Now().UnixMilli(), id)
}

func UpdateByKey(ctx context.Context, key string, alloc *Alloc) error {
	if alloc == nil {
		return ErrNilAlloc
	}

	statement := fmt.Sprintf(
		"update %s set cur_id = ?, step = ?, description = ?, updated_at = ? where biz_key = ?", TableName,
	)

	return stmtExec(ctx, statement, alloc.CurId, alloc.Step, alloc.Step, time.Now().UnixMilli(), key)
}

func UpdateCurIdById(ctx context.Context, id int64, newMaxId int64) error {
	statement := fmt.Sprintf(
		"update %s set cur_id = ? updated_at where id = ?", TableName,
	)

	return stmtExec(ctx, statement, newMaxId, id)
}

func UpdateCurIdByKey(ctx context.Context, key string, newMaxId int64) error {
	statement := fmt.Sprintf(
		"update %s set cur_id = ? updated_at = ? where biz_key = ?", TableName,
	)

	return stmtExec(ctx, statement, newMaxId, time.Now().UnixMilli(), key)
}

// return curId before update
func UpdateCurIdWithStepByKey(ctx context.Context, key string) (uint64, uint32, error) {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("dao begin tx err: %v\n", err)
		return 0, 0, pkg.ErrDb.Message(err.Error())
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
		fmt.Sprintf("select cur_id, step from %s where biz_key = ? limit 1", TableName),
		key,
	)
	if err != nil {
		log.Printf("dao tx query err: %v\n", err)
		return 0, 0, pkg.ErrDb.Message(err.Error())
	}
	defer row.Close()

	var (
		curId uint64
		step  uint32
	)
	for row.Next() {
		err = row.Scan(&curId, &step)
		if err != nil {
			log.Printf("dao scan row err: %v\n", err)
			return 0, 0, pkg.ErrDb.Message(err.Error())
		}
		break
	}

	if err := row.Err(); err != nil {
		log.Printf("dao row err: %v\n", err)
		return 0, 0, pkg.ErrDb.Message(err.Error())
	}

	// update cur_id = cur_id + step
	err = txStmtExec(ctx, tx,
		fmt.Sprintf("update %s set cur_id = ? where biz_key = ?", TableName),
		curId+uint64(step),
		key,
	)
	if err != nil {
		log.Printf("dao tx stmt exec err: %v\n", err)
		return 0, 0, pkg.ErrDb.Message(err.Error())
	}

	rollback = false
	return curId, step, nil
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
