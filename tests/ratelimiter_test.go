package tests

//запуск тестов:
//go test -v -run RateLimiter

import (
	"fmt"
	"github.com/muhserg/ratelimiter/ratelimiter"
	"strconv"
	"sync"
	"testing"
)

const TASK_COUNT = 100

func TestRateLimiter(t *testing.T) {
	var wg sync.WaitGroup
	chanTask := make(chan string)

	wg.Add(1)
	go ratelimiter.LimitTasks(&wg, chanTask)

	//отправка задач
	for i := 0; i < TASK_COUNT; i++ {
		task := `{"name": "Задача #` + strconv.Itoa(i+1) + `", "text": "Сделать задачу"}`
		chanTask <- task
	}

	//закрываем канал, чтобы горутина LimitTasks завершила работу
	close(chanTask)
	wg.Wait()
	fmt.Println("Задачи отправлены")
}
