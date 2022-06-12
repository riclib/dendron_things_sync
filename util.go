package main

import (
	"log"
	"strconv"
	"time"
)

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

