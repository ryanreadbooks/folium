package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

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
}

func CloseDB() {
	if db != nil {
		err := db.Close()
		if err != nil {
			log.Printf("can not close db: %v", err)
		}
	}
}

func GetDB() *sql.DB {
	return db
}

func getDbDsn(user, pass, addr, dbName string) string {
	// [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parsetTime=True&loc=Local",
		user, pass, addr, dbName)
}

// segment table
const (
	allocTableName = "alloc_table"
)

type Alloc struct {
	Id        int64  // primary key
	Key       string // unique key
	MaxId     uint64
	Step      uint64
	Desc      string
	CreatedAt uint64
	UpdatedAt uint64
}

var (
	ErrNilAlloc = ErrInvalidArgs.Message("alloc arg is nil")
)

func CreateAlloc(ctx context.Context, alloc *Alloc) (int64, error) {
	if alloc == nil {
		return 0, ErrNilAlloc
	}

	query := fmt.Sprintf(
		"insert into %s(key,max_id,step,desc,created_at,updated_at) values (?,?,?,?,?,?)", allocTableName,
	)
	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("prepare stmt err: %v", err)
		return 0, ErrDb.Message(err.Error())
	}
	defer stmt.Close()

	ctime := time.Now().UnixMilli()
	res, err := stmt.ExecContext(ctx, alloc.Key, alloc.MaxId, alloc.Step, alloc.Desc, ctime, ctime)
	if err != nil {
		log.Printf("stmt exec err: %v", err)
		return 0, ErrDb.Message(err.Error())
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		log.Printf("get last inserted id err: %v", err)
		return 0, ErrDb.Message(err.Error())
	}

	return lastId, nil
}

func stmtExec(ctx context.Context, statement string, args ...interface{}) error {
	stmt, err := db.PrepareContext(ctx, statement)
	if err != nil {
		log.Printf("prepare stmt err: %v", err)
		return ErrDb.Message(err.Error())
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, args...)
	if err != nil {
		log.Printf("stmt exec err: %v", err)
		return ErrDb.Message(err.Error())
	}

	return nil
}

func UpdateAllocById(ctx context.Context, id int64, alloc *Alloc) error {
	if alloc == nil || id == 0 {
		return ErrInvalidArgs.Message("invalid args when update")
	}

	statement := fmt.Sprintf(
		"update %s set max_id = ?, step = ?, desc = ?, updated_at = ? where id = ?", allocTableName,
	)

	return stmtExec(ctx, statement, alloc.MaxId, alloc.Step, alloc.Step, time.Now().UnixMilli(), id)
}

func UpdateAllocByKey(ctx context.Context, key string, alloc *Alloc) error {
	if alloc == nil {
		return ErrNilAlloc
	}

	statement := fmt.Sprintf(
		"update %s set max_id = ?, step = ?, desc = ?, updated_at = ? where key = ?", allocTableName,
	)

	return stmtExec(ctx, statement, alloc.MaxId, alloc.Step, alloc.Step, time.Now().UnixMilli(), key)
}
