package dal

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

var once sync.Once
var db *sql.DB

func Connect() (*sql.DB, error) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}
	dbname := os.Getenv("DBNAME")
	user := os.Getenv("USER")
	userpassword := os.Getenv("USERPASSWORD")
	if dbname == "" || user == "" || userpassword == "" {
		fmt.Println("error in data fetching from env file")
	}
	// var err error
	once.Do(func() {
		connection_string := "postgresql://" + user + ":" + userpassword + "@solar-ape-6502.8nk.cockroachlabs.cloud:26257/" + dbname + "?sslmode=verify-full"
		// fmt.Println(connection_string)
		db, err = sql.Open("postgres", connection_string)
		if err != nil {
			fmt.Println("Database Connection err", err)
			return
		}

	})
	return db, err
}

func MustExec(query string, args ...interface{}) (int64, error) {
	db = GetDB()
	result, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	RowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println("RowsAffected Error", err)
		return 0, err
	}
	return RowsAffected, err

}
func GetDB() *sql.DB {
	return db
}
