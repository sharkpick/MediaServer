package main

import (
	"database/sql"
)

func createSessionTable(db *sql.DB) {
	query := `CREATE TABLE ` + MyConfig.SessionTable + `(
		"id" TEXT NOT NULL PRIMARY KEY,
		"ip" TEXT NOT NULL);`
	statement, err := db.Prepare(query)
	if err != nil {
		//log.Println("error in createSessionTable:", err)
		return
	}
	statement.Exec()
}

func createVideoTable(db *sql.DB) {
	query := `CREATE TABLE ` + MyConfig.VideoTable + `(
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"title" TEXT NOT NULL UNIQUE,
		"path" TEXT NOT NULL UNIQUE,
		"format" TEXT NOT NULL,
		"size" integer NOT NULL,
		"adult" integer NOT NULL);`
	statement, err := db.Prepare(query)
	if err != nil {
		return
	}
	statement.Exec()
}
