package ratelimiter

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/logs"
	"strconv"
	"sync"
	"time"
)

var config map[string]int

type Task struct {
	Name string
	Text string
}

const END_TASK string = "end"
const CHECK_RETRY_TIMEOUT time.Duration = 100

//Настройка ограничений
func init() {
	config = make(map[string]int)
	config["maxSimultaneouslyTask"] = 5
	config["maxTaskInMinute"] = 40
}

//Запуск задач в соответствии с ограничениями
func LimitTasks(wg *sync.WaitGroup, chanTask <-chan string) {
	defer wg.Done()

	chanResultTaskEnd := make(chan string, config["maxSimultaneouslyTask"])
	defer close(chanResultTaskEnd)

	var beginTasks, endTasks, tasksInMinute int
	start := time.Now()
	var mutex sync.Mutex

	for {
		task, ok := <-chanTask
		if !ok {
			break
		}

		if calcExecTask(chanResultTaskEnd, &beginTasks, &endTasks, &tasksInMinute, &start) {
			go taskChanExec(chanResultTaskEnd, task, &mutex)
			beginTasks++
			tasksInMinute++
		}
	}

	fmt.Println("Задачи отправлены параллельно")
}

//Запуск конкретной задачи с уведомлением о завершении (chanResultTaskEnd только для записи)
func taskChanExec(chanResultTaskEnd chan<- string, task string, mutex *sync.Mutex) {

	err := taskExec(task)
	if err != nil {
		logs.Error(err)
	}

	mutex.Lock()
	chanResultTaskEnd <- END_TASK
	mutex.Unlock()
}

//Запуск конкретной задачи
func taskExec(task string) error {
	var taskObj Task
	err := json.Unmarshal([]byte(task), &taskObj)
	if err != nil {
		return err
	}

	fmt.Printf("В задаче '%s' нужно сделать следующее: %s\n", taskObj.Name, taskObj.Text)
	time.Sleep(3 * time.Second)

	return nil
}

//Получает сообщения о выполнении задач из канала
func getTaskResultMess(chanResultTaskEnd <-chan string, endTasks *int) bool {
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

//Вычисляет число запущенных задач и выдает сигнал запускать новые задачи
func calcExecTask(chanResultTaskEnd <-chan string, beginTasks, endTasks, tasksInMinute *int, start *time.Time) bool {
	for {
		chanOpen := getTaskResultMess(chanResultTaskEnd, endTasks)
		if !chanOpen {
			logs.Error(errors.New("Канал результатов задач chanResultTaskEnd преждевременно закрылся."))
			break
		}

		diff := time.Minute - time.Since(*start)
		if diff < 0 { //когда задачи выполняются медленно
			*tasksInMinute = 0
		} else if *tasksInMinute == config["maxTaskInMinute"] { //когда задачи выполняются быстро
			*tasksInMinute = 0
			fmt.Println("Wait " + strconv.FormatFloat(diff.Seconds(), 'f', 2, 64) + " seconds...")
			time.Sleep(diff)
			*start = time.Now()
		}

		if *beginTasks-*endTasks < config["maxSimultaneouslyTask"] {
			return true
		}
		time.Sleep(CHECK_RETRY_TIMEOUT * time.Millisecond)
	}

	return false
}
