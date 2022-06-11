package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
	gm "github.com/yuin/goldmark/text"
	"gopkg.in/yaml.v2"
)

type TaskData struct {
	ID       string `yaml:"id"`
	Title    string `yaml:"title"`
	Desc     string `yaml:"desc"`
	Updated  string `yaml:"updated"`
	Created  string `yaml:"created"`
	Status   string `yaml:"status"`
	Due      string `yaml:"due"`
	Priority string `yaml:"priority"`
	Owner    string `yaml:"owner"`
	Notes    string
}



const (
	notesRoot   = "/Users/riclib/repo/notes/notes/task*.md"
	thingsToken = "vnsjaoUjTxmehnAgGHGDqg"
)

func main() {
	

	matches, err := filepath.Glob(notesRoot)
	if err != nil {
		log.Fatal("Couldn't list files", err)
	}
	for _, match := range matches {
		task, err := getTask(match)
		if err != nil {
			log.Println("couldn't parse file", "file", match, "error", err)
		}
		fmt.Println(task)
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

func getTask( path string) (TaskData, error) {
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
		return  task, errors.New("File doesn't start with ---")
	}

	//Read Metadata
	metadata, found := readUntilSeparator(scanner)
	if !found {
		return  task, errors.New("Couldn't find metadata")
	}
	err = yaml.Unmarshal([]byte(metadata), &task)
	if err != nil {
		return  task, errors.New("Couldn't unmarshall task")
	}
	err = parseDates(&task)
	if err != nil {
		return  task, errors.New("Couldn't parse dates")
	}

	task.Notes = readUntil10k(scanner)
	return task, nil
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

func readUntil10k(scanner *bufio.Scanner) (string) {
	text := ""
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) + len(text) > 10000 {
			break
		}
		text = text + line + "\n"
	}
	return text
}


func printTitles(markdown goldmark.Markdown, path string) error {

	b, err := os.ReadFile(path)
	mdreader := gm.NewReader(b)

	if err != nil {
		return err
	}

	context := parser.NewContext()
	var buf bytes.Buffer
	if err := markdown.Convert([]byte(b), &buf, parser.WithContext(context)); err != nil {
		panic(err)
	}

	mdparser := parser.NewParser()
	mdparser.Parse(mdreader, parser.WithContext(context))

	metaData := meta.Get(context)
	id := fmt.Sprint(metaData["id"])
	title := fmt.Sprint(metaData["title"])
	status, found := metaData["status"]
	if found {
		statusStr := fmt.Sprint(status)
		fmt.Println(id + "[" + statusStr + "]: " + title)
		url, err := genThingsURL(title, "No notes yet", time.Now(), "", time.Now(), time.Now(), false, false)
		if err != nil {
			log.Println("Error generating things url", err)
		}
		fmt.Println(url)
	} else {
		fmt.Println(id + ":" + title + " -> not a task")
	}

	return nil
}

func genThingsURL(title string, notes string, deadline time.Time, list string, creationDate time.Time, completionDate time.Time, completed bool, cancelled bool) (string, error) {
	path := ""
	path = addURLencoded(path, "title", title)
	path = addURLencoded(path, "notes", notes)
	path = addURLencoded(path, "auth-token", thingsToken)
	return "things:///add?" + path, nil

}

func addURLencoded(path, param, value string) string {
	encoded := url.PathEscape(value)
	if path != "" {
		path = path + "&"
	}
	return path + param + "=" + encoded
}
