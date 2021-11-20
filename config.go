package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var DefaultConfig = &Config{
	MediaDirectory:    "/srv/Media/",
	AdultDirectory:    "/srv/Media/Porn/",
	MovieDirectory:    "/srv/Media/Movies/",
	DatabaseFile:      "./VideoDatabase.db",
	VideoTable:        "tVideo",
	SessionTable:      "tSession",
	UpdateTimeSeconds: 30,
	ListenPort:        ":2323",
	SettingsFile:      "settings.json",
}

type ConfigManager struct {
	*Config
	done  chan interface{}
	mutex sync.Mutex
}

type Config struct {
	MediaDirectory    string
	DatabaseFile      string
	AdultDirectory    string
	MovieDirectory    string
	VideoTable        string
	SessionTable      string
	ListenPort        string
	UpdateTimeSeconds int
	SettingsFile      string
}

func (c *ConfigManager) FlushSettings() {
	file, _ := json.MarshalIndent(c.Config, "", " ")
	ioutil.WriteFile(c.SettingsFile, file, 0644)
}

var SettingsLastModified time.Time

func (c *ConfigManager) ReloadConfig() {
	for {
		select {
		case <-c.done:
			return
		default:
			c.mutex.Lock()
			if info, err := os.Stat(c.SettingsFile); err != nil {
				// settings file gone - flush
				c.FlushSettings()
				if _, err := os.Stat(c.SettingsFile); err != nil {
					log.Fatalln("error -", c.SettingsFile, "not found", err)
				}
			} else {
				if info.ModTime().After(SettingsLastModified) {
					SettingsLastModified = info.ModTime()
					log.Println("found", c.SettingsFile, "changed - updating")
					tmp := ReadConfig(c.SettingsFile)
					c.Config = nil
					c.Config = tmp
				}
			}
			c.mutex.Unlock()
			time.Sleep(time.Second)
		}
	}
}

func ReadConfig(path string) *Config {
	var c *Config
	if f, err := os.ReadFile(path); err == nil {
		json.Unmarshal(f, &c)
		if false == strings.HasPrefix(c.ListenPort, ":") {
			c.ListenPort = ":" + c.ListenPort
		}
		if "" == c.SettingsFile {
			c.SettingsFile = "settings.json"
		}
	}
	return c
}

func NewConfig() *ConfigManager {
	m := ConfigManager{}
	if c := ReadConfig("settings.json"); c == nil {
		log.Println("making new config")
		m.Config = DefaultConfig
		m.FlushSettings()
	} else {
		m.Config = c
	}
	m.done = make(chan interface{})
	if info, err := os.Stat(m.SettingsFile); err != nil {
		log.Fatalln("Error -", m.SettingsFile, "not found")
	} else {
		SettingsLastModified = info.ModTime()
	}
	go m.ReloadConfig()
	return &m
}
