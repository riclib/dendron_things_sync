package main

import (
	"log"
	"regexp"
)

type TaskData struct {
	ID       string  `yaml:"id"`
	Title    string  `yaml:"title"`
	Desc     string  `yaml:"desc"`
	Updated  string  `yaml:"updated"`
	Created  string  `yaml:"created"`
	Status   *string `yaml:"status"`
	Due      string  `yaml:"due"`
	Priority string  `yaml:"priority"`
	Owner    string  `yaml:"owner"`
	ThingsID string  `yaml:"things_id`
	Notes    string
	Filepath string
}

const (
	dendronDault         = "/Users/riclib/repo/notes/notes/"
	notesRoot            = "/Users/riclib/repo/notes/notes/task*.md"
	thingsToken          = "vnsjaoUjTxmehnAgGHGDqg"
	WaitTimeSecs         = 10
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
	Watcher()
}
