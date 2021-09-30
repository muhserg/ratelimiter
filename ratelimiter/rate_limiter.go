package ratelimiter

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/logs"
	"time"
)

var config map[string]int

type Task struct {
	Name string
	Text string
}

const BEGIN_TASK = "begin"
const END_TASK = "end"

//Настройка ограничений
func init() {
	config = make(map[string]int)
	config["maxSimultaneouslyTask"] = 5
	config["maxTaskInMinute"] = 10
}

//Запуск задач в соответствии с ограничениями
func LimitTasks(chanTask chan string) {
	chanResultTaskBegin := make(chan string)
	chanResultTaskEnd := make(chan string)
	//defer close(chanTask)

	var beginTasks int
	var endTasks int
	var quantitySend int
	for {
		task, ok := <-chanTask
		if !ok {
			break
		}

		if quantitySend == 0 {
			go taskExec(chanResultTaskBegin, chanResultTaskEnd, task)
		} else if calcExecTask(chanResultTaskBegin, chanResultTaskEnd, &beginTasks, &endTasks) {
			logs.Debug("Send task", task)
			go taskExec(chanResultTaskBegin, chanResultTaskEnd, task)
		}
		quantitySend++
	}
}

//Запуск конкретной задачи
func taskExec(chanResultTaskBegin, chanResultTaskEnd chan string, task string) {
	chanResultTaskBegin <- BEGIN_TASK
	//logs.Debug("Exec task", task)

	var taskObj Task
	err := json.Unmarshal([]byte(task), &taskObj)
	if err != nil {
		logs.Error(err)
		chanResultTaskEnd <- END_TASK
		return
	}

	fmt.Println("В задаче '" + taskObj.Name + "' нужно сделать следующее: " + taskObj.Text)
	time.Sleep(4 * time.Second)

	chanResultTaskEnd <- END_TASK
}

//Получает сообщения о выполнении задач из канала
func getTaskResultMess(chanResultTaskBegin, chanResultTaskEnd chan string, beginTasks, endTasks *int) bool {
	ok := false

	select {
		case _, ok = <-chanResultTaskBegin:
			*beginTasks++
		case _, ok = <-chanResultTaskEnd:
			*endTasks++
	}

	return ok
}

//Вычисляет число запущенных задач и выдает сигнал запускать новые задачи
func calcExecTask(chanResultTaskBegin, chanResultTaskEnd chan string, beginTasks, endTasks *int) bool {
	for {
		chanOpen := getTaskResultMess(chanResultTaskBegin, chanResultTaskEnd, beginTasks, endTasks)
		if !chanOpen {
			logs.Error(errors.New("Канал результатов задач преждевременно закрылся."))
			break
		}
		logs.Debug("beginTasks: ", *beginTasks, ", endTasks: ", *endTasks)

		if *beginTasks-*endTasks < config["maxSimultaneouslyTask"] {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}

	return false
}
