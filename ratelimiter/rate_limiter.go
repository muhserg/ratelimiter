package ratelimiter

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/logs"
	"sync"
	"time"
)

var config map[string]int

type Task struct {
	Name string
	Text string
}

const END_TASK = "end"
const CHECK_RETRY_TIMEOUT = 100

//Настройка ограничений
func init() {
	config = make(map[string]int)
	config["maxSimultaneouslyTask"] = 5
	config["maxTaskInMinute"] = 10
}

//Запуск задач в соответствии с ограничениями
func LimitTasks(wg *sync.WaitGroup, chanTask chan string) {
	defer wg.Done()

	chanResultTaskEnd := make(chan string, config["maxSimultaneouslyTask"])
	defer close(chanResultTaskEnd)

	var beginTasks int
	var endTasks int
	for {
		task, ok := <-chanTask
		if !ok {
			break
		}

		if calcExecTask(chanResultTaskEnd, &beginTasks, &endTasks) {
			//logs.Debug("Send task", task)
			go taskChanExec(chanResultTaskEnd, task)
			beginTasks++
		}
	}

	fmt.Println("Задачи отправлены параллельно")
}

//Запуск конкретной задачи  уведомлением о завершении
func taskChanExec(chanResultTaskEnd chan string, task string) {


	err := taskExec(task)
	if err != nil {
		logs.Error(err)
		chanResultTaskEnd <- END_TASK
	}

	chanResultTaskEnd <- END_TASK
}

//Запуск конкретной задачи
func taskExec(task string) error {
	var taskObj Task
	err := json.Unmarshal([]byte(task), &taskObj)
	if err != nil {
		return err
	}

	fmt.Println("В задаче '" + taskObj.Name + "' нужно сделать следующее: " + taskObj.Text)
	time.Sleep(3 * time.Second)

	return nil
}

//Получает сообщения о выполнении задач из канала
func getTaskResultMess(chanResultTaskEnd chan string, endTasks *int) bool {
	select {
	case _, ok := <-chanResultTaskEnd:
		if !ok {
			return false
		}
		*endTasks++
	default:
	}

	return true
}

//очищаем канал от сообщений END_TASK
func clearTaskResultMess(chanResultTaskEnd chan string) {
	answer := "ready"
	for  {
		select {
		case _, ok := <-chanResultTaskEnd:
			if !ok {
				return
			}
		default:
			answer = "empty"
		}

		if answer == "empty" {
			break
		}
	}
}

//Вычисляет число запущенных задач и выдает сигнал запускать новые задачи
func calcExecTask(chanResultTaskEnd chan string, beginTasks, endTasks *int) bool {
	for {
		chanOpen := getTaskResultMess(chanResultTaskEnd, endTasks)
		if !chanOpen {
			logs.Error(errors.New("Один из двух каналов результатов задач преждевременно закрылся."))
			break
		}

		//logs.Debug("beginTasks: ", *beginTasks, ", endTasks: ", *endTasks)
		if *beginTasks-*endTasks < config["maxSimultaneouslyTask"] {
			return true
		}
		time.Sleep(CHECK_RETRY_TIMEOUT * time.Millisecond)
	}

	return false
}
