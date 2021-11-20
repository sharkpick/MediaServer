package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	theManager *SessionManager
	theMutex   sync.Mutex
)

type SessionManager struct {
	db *sql.DB
}

type Session struct {
	ID, IP string
}

func NewSessionManager(db *sql.DB) *SessionManager {
	if theManager == nil {
		theMutex.Lock()
		defer theMutex.Unlock()
		if theManager == nil {
			theManager = &SessionManager{db: db}
		}
	}
	return theManager
}

func (s *SessionManager) SaveSession(session Session) {
	insertQuery := `INSERT OR REPLACE INTO ` + SessionTable + `(id, ip) VALUES (?, ?)`
	statement, err := s.db.Prepare(insertQuery)
	if err != nil {
		log.Fatalln("error saving session", err)
	}
	_, err = statement.Exec(session.ID, session.IP)
	if err != nil {
		log.Println("error saving session", err)
	}
}

func (s *SessionManager) NewSession(ip string) Session {
	selectQuery := `SELECT id, ip FROM ` + SessionTable + ` WHERE ip=?`
	statement, err := s.db.Prepare(selectQuery)
	if err != nil {
		log.Fatalln("Error preparing statement in NewSession:", err)
	}
	var session Session
	err = statement.QueryRow(ip).Scan(&session.ID, &session.IP)
	if err != nil {
		log.Println("warning:", err)
		session.ID = fmt.Sprintf("%016X", rand.Uint64())
		session.IP = ip
		s.SaveSession(session)
	}
	return session
}
