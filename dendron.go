package main

import (
	"bytes"
	"crypto/md5"
	"errors"
	"log"
	"net/url"
	"os"

	"github.com/natefinch/atomic"
	"gopkg.in/yaml.v2"
)

var ourChanges map[string][16]byte = make(map[string][16]byte)
var sep = []byte("---\n")

func getTask(path string) (t TaskData, wasTask bool, err error) {
	var task TaskData

	data, err := os.ReadFile(path)
	if err != nil {
		return task, false, errors.New("couldn't read file")
	}
	hash := md5.Sum(data)
	if ourChanges[path] == hash {
		log.Println("ignoring our own change", path)
		return task, false, nil
	}

	parts := bytes.SplitAfterN(data, sep, 4)
	if len(parts) != 3 {
		return t, false, errors.New("note without metadata")
	}
	metadata := parts[1]
	task.Contents = parts[2]

	err = yaml.Unmarshal(metadata, &task)
	if err != nil {
		return task, false, errors.New("couldn't unmarshall task")
	}
	if task.Status == nil { //not a task
		return task, false, nil
	}
	task.Filepath = path
	thingsFooter := "\n" + string(sep) + "dendron_id: " + task.ID + "\n" + "filepath: " + url.PathEscape("file:/" + task.Filepath) + "\n"
	thingsNotesSpaceLeft := 10000 - len(thingsFooter)
	if thingsNotesSpaceLeft >= len(task.Contents) {
		task.ThingsNotes = string(task.Contents) + thingsFooter  
	} else {
		task.ThingsNotes = task.ThingsNotes + string(task.Contents[0:thingsNotesSpaceLeft-1]) + thingsFooter
	}
	return task, true, nil
}

func putTask(task TaskData) error {
	result := sep
	header, err := yaml.Marshal(task)
	if err != nil {
		return err
	}
	result = append(result, header...)
	//	result = append(result, []byte("\n")...)
	result = append(result, sep...)
	result = append(result, task.Contents...)
	err = atomic.WriteFile(task.Filepath, bytes.NewReader(result))
	ourChanges[task.Filepath] = md5.Sum(result)
	return err
}
