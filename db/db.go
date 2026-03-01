package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func createDbFolder() error {
	if _, err := os.Stat("data"); os.IsNotExist(err) {
		if err := os.Mkdir("data", 0755); err != nil {
			return err
		}
		fmt.Println("Data folder created successfully")
	} else {
		return err
	}
	return nil

}

func ConnectToDb() {
	if err := createDbFolder(); err != nil {
		panic(err)
	}

	connection, err := sql.Open("sqlite3", "./data/test.db?_journal_mode=WAL")
	if err != nil {
		panic(err)
	}
	db = connection
}

func GetDb() *sql.DB {
	return db
}

func SeedTables() {
	db.Exec(`CREATE TABLE IF NOT EXISTS jobs(id integer primary key, title varchar(50) NOT NULL,
	 	endpoint varchar(50) NOT NULL, method varchar(6) NOT NULL, payload TEXT NOT NULL,
		scheduled_at INTEGER NOT NULL, system_scheduled_at INTEGER NOT NULL, locked_at INTEGER, created_on TEXT NOT NULL DEFAULT(datetime('now')),
		status VARCHAR(10) NOT NULL CHECK (status IN ('IDLE', 'RUNNING','SUCCESS', 'FAILED', 'ABORTED' )) DEFAULT 'IDLE',
		retries SMALLINT, error_info TEXT, updated_on TEXT)
		`)
}
