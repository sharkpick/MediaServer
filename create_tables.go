package main

import (
	"database/sql"
	"log"
)

func createSessionTable(db *sql.DB) {
	query := `CREATE TABLE ` + SessionTable + `(
		"id" TEXT NOT NULL PRIMARY KEY,
		"ip" TEXT NOT NULL);`
	statement, err := db.Prepare(query)
	if err != nil {
		log.Println("error in createSessionTable:", err)
		return
	}
	_, err = statement.Exec()
	if err != nil {
		log.Println("error in createSessionTable:", err)
	}
}

func createVideoTable(db *sql.DB) {
	query := `CREATE TABLE ` + VideoTable + `(
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"title" TEXT NOT NULL UNIQUE,
		"path" TEXT NOT NULL UNIQUE,
		"format" TEXT NOT NULL,
		"size" integer NOT NULL,
		"adult" integer NOT NULL);`
	statement, err := db.Prepare(query)
	if err != nil {
		log.Println("error in createSessionTable:", err)
		return
	}
	_, err = statement.Exec()
	if err != nil {
		log.Println("error in createSessionTable:", err)
	}
}
