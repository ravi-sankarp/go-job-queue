package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func ConnectToDb() {
	connection, err := sql.Open("sqlite3", "./test.db")
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
		scheduled_at TEXT NOT NULL, created_on TEXT NOT NULL DEFAULT(datetime('now')),
		status VARCHAR(10) NOT NULL CHECK (status IN ('IDLE', 'RUNNING','SUCCESS', 'FAILED' )) DEFAULT 'IDLE',
		retries SMALLINT, error_info TEXT, updated_on TEXT)
		`)
}
