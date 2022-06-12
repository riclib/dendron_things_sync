package main

import (
	"log"
	"regexp"
)

type TaskData struct {
	ID          string      `yaml:"id"`
	Title       string      `yaml:"title"`
	Desc        string      `yaml:"desc"`
	Updated     interface{} `yaml:"updated"`
	Created     interface{} `yaml:"created"`
	Status      *string     `yaml:"status"`
	Due         interface{} `yaml:"due"`
	Priority    string      `yaml:"priority"`
	Owner       string      `yaml:"owner"`
	ThingsID    string      `yaml:"_things_id"`
	ThingsNotes string      `yaml:"-"`
	Contents    []byte      `yaml:"-"`
	Filepath    string      `yaml:"-"`
	TODO        string      `yaml:"TODO"`
}

const (
	dendronDault         = "/Users/riclib/repo/notes/notes/"
	notesRoot            = "/Users/riclib/repo/notes/notes/task*.md"
	thingsToken          = "vnsjaoUjTxmehnAgGHGDqg"
	WaitTimeSecs         = 3
	MaxWaitForThingsSecs = 10.0
)

var dendronFilters = []string{`.+task\..+\..+\..+\..+\.md`}
var filterRegexes []regexp.Regexp

func main() {
	err := initThingsConnection()
	if err != nil {
		log.Fatal("couldn't connect to things db")
	}
	for _, f := range dendronFilters {
		filterRegexes = append(filterRegexes, *regexp.MustCompile(f))
	}
	InitialSync()
	Watcher()
}
