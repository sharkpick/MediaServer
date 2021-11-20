package main

import (
	"database/sql"
	"log"
	"net"
	"net/http"
	"strconv"
	"text/template"

	_ "github.com/mattn/go-sqlite3"
)

const (
	MediaDirectory    = "/srv/Media/"
	AdultDirectory    = MediaDirectory + "/Porn/"
	MovieDirectory    = MediaDirectory + "/Movies/"
	DatabaseFile      = "./VideoDatabase.db"
	VideoTable        = "tVideo"
	SessionTable      = "tSession"
	UpdateTimeMinutes = 15
)

var (
	MyCollection *MediaCenter
	MySessions   *SessionManager
)

func serveVideo(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	s := MySessions.NewSession(ip)
	log.Println(s.ID, s.IP, r.URL.String())
	var video *Video
	var err error
	if len(r.URL.Query()["title"]) > 0 {
		title := r.URL.Query()["title"][0]
		video, err = MyCollection.getVideoByTitle(title)
	} else if len(r.URL.Query()["videoid"]) > 0 {
		videoIDString := r.URL.Query()["videoid"][0]
		videoID, _ := strconv.Atoi(videoIDString)
		video, err = MyCollection.getVideoByID(videoID)
	} else {
		http.Error(w, "bad request", http.StatusBadRequest)
	}
	if err != nil {
		http.Error(w, "video not found", http.StatusNotFound)
	} else {
		http.ServeFile(w, r, video.Path)
	}
}

func doIndex(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	s := MySessions.NewSession(ip)
	log.Println(s.ID, s.IP, r.URL.String())
	if t, err := template.ParseFiles("index.html"); err != nil {
		log.Fatalln("Error: can't serve index.html", err)
	} else {
		adult, general := MyCollection.Videos()
		t.Execute(w, struct {
			Adult   []Video
			General []Video
		}{
			Adult:   adult,
			General: general,
		})
	}
}

func setupHandlers() {
	http.HandleFunc("/", doIndex)
	http.HandleFunc("/video", serveVideo)
}

func main() {
	setupHandlers()
	db, err := sql.Open("sqlite3", DatabaseFile)
	if err != nil {
		log.Fatalln("Error opening database:", err)
	}
	defer db.Close()
	createSessionTable(db)
	createVideoTable(db)
	MyCollection = NewMediaCenter(db)
	MySessions = NewSessionManager(db)
	log.Fatalln(http.ListenAndServe(":1990", nil))
}
