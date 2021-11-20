package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	theMediaCenter *MediaCenter
)

type Video struct {
	ID                  int
	Size                int64
	Path, Format, Title string
	Adult               bool
}

func (v *Video) Print() string {
	return fmt.Sprintf("%d %s", v.ID, v.Path)
}

func NewVideo(path string) (*Video, error) {
	v := Video{}
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("file not found", err)
	}
	v.Path = path
	if strings.Contains(path, AdultDirectory) {
		v.Adult = true
	} else {
		v.Adult = false
	}
	v.Size = info.Size()
	v.Title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	v.Format = strings.ReplaceAll(filepath.Ext(path), ".", "")
	return &v, nil
}

type MediaCenter struct {
	db   *sql.DB
	done chan interface{}
}

func (m *MediaCenter) Videos() (a, g []Video) {
	var adult []Video
	var general []Video
	selectQuery := `SELECT id, size, path, format, title, adult FROM ` + VideoTable + ` ORDER BY title ASC`
	statement, err := m.db.Prepare(selectQuery)
	if err != nil {
		log.Fatalln("Fatal Error:", err)
	}
	row, err := statement.Query()
	if err != nil {
		log.Println("Error in m.Videos():", err)
	}
	defer row.Close()
	for row.Next() {
		var video Video
		var ad int
		err = row.Scan(&video.ID, &video.Size, &video.Path, &video.Format, &video.Title, &ad)
		if err != nil {
			log.Println("Error in m.Videos():", err)
			continue
		}
		video.Adult = ad == 1
		if video.Adult {
			adult = append(adult, video)
		} else {
			general = append(general, video)
		}

	}
	return adult, general
}

func (m *MediaCenter) Add(v *Video, id ...int) {
	insertSQL := `INSERT INTO ` + VideoTable + `(path, title, format, size, adult) VALUES (?, ?, ?, ?, ?)`
	statement, err := m.db.Prepare(insertSQL)
	if err != nil {
		log.Fatalln("Fatal Error:", err)
	}
	adult := func() int {
		if v.Adult {
			return 1
		}
		return 0
	}()
	_, err = statement.Exec(v.Path, v.Title, v.Format, v.Size, adult)
	if err != nil {
		log.Printf("Error in Add(%s): %s", v.Title, err)
	}
}

func (m *MediaCenter) monitorForNewFiles(done <-chan interface{}) {
	for {
		select {
		case <-m.done:
			return
		default:
			for _, dir := range []string{MovieDirectory, AdultDirectory} {
				files, err := ioutil.ReadDir(dir)
				if err != nil {
					log.Fatalln("Error walking", dir, err)
				}
				for _, file := range files {
					path := dir + file.Name()
					if _, err := m.getVideoByTitle(file.Name()); err != nil {
						if ".mp4" == filepath.Ext(path) || ".m4v" == filepath.Ext(path) {
							v, err := NewVideo(path)
							if err != nil {
								log.Println(err)
								continue
							}
							m.Add(v)
						} else {
							log.Println("invalid file format", filepath.Ext(path))
						}
					}
				}
			}

			time.Sleep(time.Minute * UpdateTimeMinutes)
		}
	}
}

func NewMediaCenter(db *sql.DB) *MediaCenter {
	if theMediaCenter == nil {
		theMutex.Lock()
		defer theMutex.Unlock()
		if theMediaCenter == nil {
			theMediaCenter = &MediaCenter{db: db, done: make(chan interface{})}
			go theMediaCenter.monitorForNewFiles(theMediaCenter.done)
		}
	}
	return theMediaCenter
}

func (m *MediaCenter) getVideoByTitle(title string) (*Video, error) {
	selectQuery := `SELECT id, size, path, format, title FROM ` + VideoTable + ` WHERE title=?`
	statement, err := m.db.Prepare(selectQuery)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	var video Video
	err = statement.QueryRow(title).Scan(&video.ID, &video.Size, &video.Path, &video.Format, &video.Title)
	return &video, err
}

func (m *MediaCenter) getVideoByID(videoID int) (*Video, error) {
	selectQuery := `SELECT id, size, path, format, title FROM ` + VideoTable + ` WHERE id=?`
	statement, err := m.db.Prepare(selectQuery)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	var video Video
	err = statement.QueryRow(videoID).Scan(&video.ID, &video.Size, &video.Path, &video.Format, &video.Title)
	return &video, err
}
