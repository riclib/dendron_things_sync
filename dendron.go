package main

import (
	"bufio"
	"errors"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

func getTask(path string) (t TaskData, wasTask bool, err error) {
	var task TaskData

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Read and Check first Line
	_, found := readUntilSeparator(scanner)
	if !found {
		return task, true, errors.New("file doesn't start with ---")
	}

	//Read Metadata
	metadata, found := readUntilSeparator(scanner)
	if !found {
		return task, false, errors.New("couldn't find metadata")
	}
	err = yaml.Unmarshal([]byte(metadata), &task)
	if err != nil {
		return task, false, errors.New("couldn't unmarshall task")
	}
	if task.Status == nil { //not a task
		return task, false, nil
	}
	err = parseDates(&task)
	if err != nil {
		return task, false, errors.New("couldn't parse dates")
	}
	task.Notes = "dendron_id: " + task.ID + "\n" + "filepath: file:" + task.Filepath + "\n"
	task.Notes = task.Notes + readUntilMaxSize(scanner, 10000-len(task.Notes))
	return task, true, nil
}

// reads the file data until the next separator
// returns false if separator wasn't found
func readUntilSeparator(scanner *bufio.Scanner) (string, bool) {
	text := ""
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			return text, true
		}
		text = text + line + "\n"
	}
	return "", false
}

func readUntilMaxSize(scanner *bufio.Scanner, maxsize int) string {
	text := ""
	for scanner.Scan() {
		line := scanner.Text()
		if len(line)+len(text) > maxsize {
			break
		}
		text = text + line + "\n"
	}
	return text
}