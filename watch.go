package main

import (
	"log"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

var ChangedFiles map[string]bool = make(map[string]bool)
var ChangedFilesMutex sync.Mutex

//if the conditions are met, execute the shell script
func InitialSync() {
	matches, err := filepath.Glob(dendronDault + "task*.md")
	if err != nil {
		log.Fatal("couldn't open dendron vault")
	}
	for _, file := range matches {
		ChangedFiles[file] = true
	}
	runSync()
}

func runSync() {
	log.Println("Running Sync")
	log.Println(ChangedFiles)
	ChangedFilesMutex.Lock()
	filesToProcess := ChangedFiles
	ChangedFiles = make(map[string]bool)
	ChangedFilesMutex.Unlock()

	//	start := time.Now()
	syncedTasks := []TaskData{}
	for filePath := range filesToProcess {
		for _, m := range filterRegexes {
			if m.Match([]byte(filePath)) {
				log.Println("processing ", filePath)
				synced, task := SyncToThings(filePath)
				if synced {
					syncedTasks = append(syncedTasks, task)
				}
			} else {
				log.Println("skipped ", filePath)
			}

		}
	}
	waitforThings := 0.1
	for _, task := range syncedTasks {
		if task.ThingsID == "" {

			var found bool
			var id string
			//		var updated int64
			for waitforThings < MaxWaitForThingsSecs {
				found, id, _ = getThingsId(task.Title, task.ID)
				if found {
					break
				} else if waitforThings < MaxWaitForThingsSecs {
					time.Sleep(time.Duration(waitforThings * float64(time.Second)))
					log.Println("Waited", time.Duration(waitforThings))
					waitforThings *= 2
				}
			}
			task.ThingsID = id
			err := putTask(task)
			if err != nil {
				log.Println("failed to update task", task.ID, task.ThingsID, task.Title, err)

			} else {
				log.Println("updated task", task.ID, task.ThingsID, task.Title)
			}
		}
	}

	//drain changes made by us
	ChangedFilesMutex.Lock()
	for _, t := range syncedTasks {
		found, _ := ChangedFiles[t.Filepath]
		if found {
			log.Println("removing notification for", t.Filepath)
		}
		delete(ChangedFiles, t.Filepath)

	}
	ChangedFilesMutex.Unlock()

}

func getUpdateTimeOfTask(task TaskData) int64 {
	switch v := task.Updated.(type) {
	case int:
		return int64(v)
	case string:
		intVar, err := strconv.ParseInt(v, 0, 64)
		if err != nil {
			return -1
		} else {
			return intVar
		}
	default:
		log.Printf("don't know about type %T!\n", v)
		return -1
	}
}

//handle folder files changed event
func watchFiles(watcher *fsnotify.Watcher, ch chan int64) {
	for {
		select {
		case ev := <-watcher.Events:
			{

				ChangedFilesMutex.Lock()
				ChangedFiles[ev.Name] = true
				ChangedFilesMutex.Unlock()
				ch <- time.Now().Unix()
			}
		case err := <-watcher.Errors:
			{
				log.Println("watcher error : ", err)
				return
			}
		}
	}
}

//if folder event met, execute the sync after WaitTimeSecs
func delayedSync(ch chan int64) {
	var timer *time.Timer
	for {
		select {
		case <-ch:
			{
				if nil != timer {
					log.Printf("reset timer")
					timer.Stop()
				}
				timer = time.NewTimer(WaitTimeSecs * time.Second)
				go func() {
					<-timer.C
					runSync()
				}()
			}
		}
	}
}

func Watcher() {
	notifyCh := make(chan int64)
	watcher, err := fsnotify.NewWatcher()
	watcher.Add(dendronDault)
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	go watchFiles(watcher, notifyCh)
	go delayedSync(notifyCh)
	select {}
}
