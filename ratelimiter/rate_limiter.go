package ratelimiter

import (
	"encoding/gob"
	"strconv"
	"time"
)

var config map[string]string

//Настройка ограничений
func init() {
	config["maxSimultaneouslyTask"] = "5"
	config["maxTaskInMinute"] = "10"
}

func LimitTasks(chanTask chan) {
    for {
    	task := <-chanTask
	}
}
