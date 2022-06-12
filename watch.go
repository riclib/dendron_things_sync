package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

var ChangedFiles map[string]bool = make(map[string]bool)
var ChangedFilesMutex sync.Mutex

//if the conditions are met, execute the shell script
func runSync() {
	log.Println("Quieted down")
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
		var found bool
		var id string
		for waitforThings < MaxWaitForThingsSecs {
			found, id = getThingsId(task.Title, task.ID)
			if found {
				break
			} else if waitforThings < MaxWaitForThingsSecs {
				time.Sleep(time.Duration(waitforThings * float64(time.Second)))
				log.Println("Waited", time.Duration(waitforThings))
				waitforThings *= 2
			}
		}
		task.ThingsID = id
		log.Println("found task", task.ID, task.ThingsID, task.Title)
	}

	//drain changes made by us
	ChangedFilesMutex.Lock()
	for _, t := range syncedTasks {
		found, _ := ChangedFiles[t.Filepath]
		if found{
			log.Println("removing notification for", t.Filepath)
		}
		delete(ChangedFiles, t.Filepath)

	}
	ChangedFilesMutex.Unlock()

}

//handle folder files changed event
func watchFiles(watcher *fsnotify.Watcher, ch chan int64) {
	for {
		select {
		case ev := <-watcher.Events:
			{
				fmt.Println("notified on ", ev.Name)
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
