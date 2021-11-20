package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"text/template"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	MyCollection *MediaCenter
	MySessions   *SessionManager
	MyConfig     *ConfigManager = NewConfig()
	mux                         = http.NewServeMux()
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

func doFavicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "favicon.ico")
}

func setupHandlers() {
	mux.HandleFunc("/", doIndex)
	mux.HandleFunc("/video", serveVideo)
	mux.HandleFunc("/favicon.ico", doFavicon)
}

func main() {
	setupHandlers()
	db, err := sql.Open("sqlite3", MyConfig.DatabaseFile)
	if err != nil {
		log.Fatalln("Error opening database:", err)
	}
	defer db.Close()
	createSessionTable(db)
	createVideoTable(db)
	MyCollection = NewMediaCenter(db)
	MySessions = NewSessionManager(db)
	server := &http.Server{
		Addr:    MyConfig.ListenPort,
		Handler: mux,
	}
	done := make(chan os.Signal)
	signal.Notify(done, os.Interrupt, syscall.SIGINT)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Println("server up")
	<-done
	log.Println("server stopped")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		log.Println("killing library worker")
		close(MyCollection.done)
		log.Println("killing config worker")
		close(MyConfig.done)
		cancel()
	}()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalln("server shutdown failed:", err)
	}
	log.Println("server exited")
}
