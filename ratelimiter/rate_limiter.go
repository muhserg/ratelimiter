package ratelimiter

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"time"
)

var config map[string]int

type Task struct {
	Name string
	Text string
}

//Настройка ограничений
func init() {
	config = make(map[string]int)
	config["maxSimultaneouslyTask"] = 5
	config["maxTaskInMinute"] = 10
}

//Запуск задач в соответствии с ограничениями
func LimitTasks(chanTask chan string) {
	for {
		task, ok := <-chanTask
		if !ok {
			break
		}


		go taskExec(task)
	}
}

//запуск конкретной задачи
func taskExec(task string) {

	var taskObj Task
	err := json.Unmarshal([]byte(task), &taskObj)
	if err != nil {
		logs.Error(err)
		return
	}

	fmt.Println("В задаче '" + taskObj.Name + "' нужно сделать следующее: " + taskObj.Text)
	time.Sleep(1 * time.Second)
}