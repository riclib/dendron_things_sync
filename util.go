package main

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

func EpochToTimeStr(d interface{}) string {
	var date string
	switch v := d.(type) {
	case int:
		date = fmt.Sprint(v)
	case string:
		date = v
	default:
		log.Printf("don't know about type %T!\n", v)
	}
	if date == "" {
		return ""
	}
	t, err := strconv.ParseInt(date, 10, 64)
	if err != nil {
		log.Println("Couldn't parse time", err)
	}
	return time.Unix(t/1000, 0).Format("2006-01-02")
}
