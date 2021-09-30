package tests

//запуск тестов:
//go test -v -run RateLimiter

import (
	"fmt"
	"github.com/muhserg/ratelimiter/ratelimiter"
	"strconv"
	"testing"
)

func TestRateLimiter(t *testing.T) {

	chanTask := make(chan string, 10)
	defer close(chanTask)

	go ratelimiter.LimitTasks(chanTask)

	//отправка задач
	for i := 0; i < 1000; i++ {
		task := `{"name":"Задача #` + strconv.Itoa(i+1) + `","text":"Сделать задачу"}`
		chanTask <- task
	}

	fmt.Println("Задачи отправлены")
}
