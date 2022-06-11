package main

import (
	"bufio"
	"errors"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
	
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
	Notes    string
	Filepath string
}

const (
	notesRoot   = "/Users/riclib/repo/notes/notes/task*.md"
	thingsToken = "vnsjaoUjTxmehnAgGHGDqg"
)

func main() {
	conn, err := sqlite.OpenConn(":memory:", sqlite.OpenReadWrite)
	if err != nil {
		return err
	}
	defer conn.Close()

	matches, err := filepath.Glob(notesRoot)
	if err != nil {
		log.Fatal("Couldn't list files", err)
	}
	for _, match := range matches {
		task, wasTask, err := getTask(match)
		task.Filepath = match
		if err != nil {
			log.Println("couldn't parse file", "file", match, "error", err)
		}
		if wasTask {
			url, err := genThingsURL(task)
			// fmt.Println(url)
			if err != nil {
				log.Println("couldn't generate things url", "err", err)
			}
			cmd := exec.Command("open", url)
			// fmt.Println(cmd.String())
			err = cmd.Run()
			if err != nil {
				panic(err)
			}

			// fmt.Println(url)
			// time.Sleep(5 * time.Second)
		}
	}

}

func parseDates(task *TaskData) error {
	task.Updated = EpochToTimeStr(task.Updated)
	task.Created = EpochToTimeStr(task.Created)
	task.Due = EpochToTimeStr(task.Due)
	return nil
}

func EpochToTimeStr(date string) string {
	if date == "" {
		return ""
	}
	t, err := strconv.ParseInt(date, 10, 64)
	if err != nil {
		log.Println("Couldn't parse time", err)
	}
	return time.Unix(t/1000, 0).Format("2006-01-02")
}

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
		return task, true, errors.New("File doesn't start with ---")
	}

	//Read Metadata
	metadata, found := readUntilSeparator(scanner)
	if !found {
		return task, false, errors.New("Couldn't find metadata")
	}
	err = yaml.Unmarshal([]byte(metadata), &task)
	if err != nil {
		return task, false, errors.New("Couldn't unmarshall task")
	}
	if task.Status == nil { //not a task
		return task, false, nil
	}
	err = parseDates(&task)
	if err != nil {
		return task, false, errors.New("Couldn't parse dates")
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

func genThingsURL(task TaskData) (string, error) {
	path := ""
	path = addURLencoded(path, "tags", "dendron")
	path = addURLencoded(path, "title", task.Title)
	path = addURLencoded(path, "notes", task.Notes)
	path = addURLencoded(path, "auth-token", thingsToken)
	if *task.Status == "x" {
		path = addURLencoded(path, "completed", "true")
	}
	if *task.Status == "-" {
		path = addURLencoded(path, "canceled", "true")
	}
	path = addURLencoded(path, "creation-date", task.Created)
	path = addURLencoded(path, "deadline", task.Due)
	// Calback URLs were not reliable
	//	path = addURLencoded(path, "x-success", "shortcuts://name=pingback&file="+task.Filepath)
	return "things:///add?" + path, nil

}

func addURLencoded(path, param, value string) string {
	encoded := url.PathEscape(value)
	if path != "" {
		path = path + "&"
	}
	return path + param + "=" + encoded
}
