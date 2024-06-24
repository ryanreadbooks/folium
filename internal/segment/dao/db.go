package dao

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
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(100)
	db.SetConnMaxLifetime(time.Minute * 3)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to connect db: %v", err))
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
