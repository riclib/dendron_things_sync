package main

import (
	"log"
	"math"
	"net/url"
	"os/exec"
	"path/filepath"
	"time"

	"zombiezen.com/go/sqlite"
)

var conn *sqlite.Conn = nil

func SyncAllToThings() {
	//	start := time.Now()

	matches, err := filepath.Glob(notesRoot)
	if err != nil {
		log.Fatal("Couldn't list files", err)
	}
	for _, match := range matches {
		SyncToThings(match)
	}
}
func SyncToThings(filepath string) (synced bool, task TaskData) {

	task, wasTask, err := getTask(filepath)
	if err != nil {
		log.Println("couldn't parse file", "file", filepath, "error", err)
	}

	if wasTask {
		task.Filepath = filepath

		_, _, thingsUpdateTime := getThingsId(task.Title, task.ID)
		dendronUpdatedTime := getUpdateTimeOfTask(task)
		if dendronUpdatedTime > thingsUpdateTime {
			url, err := genThingsURL(task)
			// fmt.Println(url)
			if err != nil {
				log.Println("couldn't generate things url", "err", err)
			}
			cmd := exec.Command("open", url)
			err = cmd.Run()
			if err != nil {
				panic(err)
			}
		}
		return true, task
	}
	return false, task
}

func initThingsConnection() error {
	var err error
	conn, err = sqlite.OpenConn("/Users/riclib/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/Things Database.thingsdatabase/main.sqlite", sqlite.OpenReadOnly)
	return err
}

func getThingsNews(start time.Time) {

	if conn == nil {
		log.Fatal("no connection to things db")
		return
	}
	timeFloat := float64(start.UnixMilli()) / 1000

	// Open an in-memory database.
	conn, err := sqlite.OpenConn("/Users/riclib/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/Things Database.thingsdatabase/main.sqlite", sqlite.OpenReadOnly)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Execute a query.
	stmt := conn.Prep("SELECT uuid, title from TMTask WHERE userModificationDate >= $modDate order by userModificationDate DESC;")
	stmt.SetFloat("$modDate", timeFloat)

	for {
		if hasRow, err := stmt.Step(); err != nil {
			// ... handle error
		} else if !hasRow {
			break
		}
		//		uid := stmt.GetText("uuid")
		//		title := stmt.GetText("title")
	}

}

func getThingsId(title, dendronId string) (found bool, id string, updated int64) {
	stmt := conn.Prep("SELECT uuid, userModificationDate from TMTask WHERE trashed = 0 and title = $title order by userModificationDate DESC;")
	stmt.SetText("$title", title)

	for {
		if hasRow, err := stmt.Step(); err != nil {
			log.Println("error getting things_id", err)
		} else if !hasRow {
			break
		}
		id := stmt.GetText("uuid")
		updatedSeconds := stmt.GetFloat("userModificationDate")
		updated = int64(math.Floor(1000 * updatedSeconds))
		return true, id, updated
	}
	return false, "", -1
}

func genThingsURL(task TaskData) (string, error) {
	path := ""
	path = addURLencoded(path, "tags", "dendron")
	path = addURLencoded(path, "title", task.Title)
	path = addURLencoded(path, "notes", task.ThingsNotes)
	path = addURLencoded(path, "auth-token", thingsToken)
	if *task.Status == "x" {
		path = addURLencoded(path, "completed", "true")
	}
	if *task.Status == "-" {
		path = addURLencoded(path, "canceled", "true")
	}
	path = addURLencoded(path, "creation-date", EpochToTimeStr(task.Created))
	path = addURLencoded(path, "deadline", EpochToTimeStr(task.Due))
	path = addURLencoded(path, "reveal", "false")

	// Calback URLs were not reliable
	//	path = addURLencoded(path, "x-success", "shortcuts://name=pingback&file="+task.Filepath)
	if task.ThingsID == "" {
		return "things:///add?" + path, nil
	} else {
		path = addURLencoded(path, "id", task.ThingsID)
		return "things:///update?" + path, nil
	}
}

func addURLencoded(path, param, value string) string {
	encoded := url.PathEscape(value)
	if path != "" {
		path = path + "&"
	}
	return path + param + "=" + encoded
}
