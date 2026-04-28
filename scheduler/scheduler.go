package scheduler

import (
	"execEngine/tasks"
	"fmt"
	"sync"
)

func AddToRetryQ(taskQ tasks.TaskQueue, retryQ tasks.RetryQueue) {
	for retryTask := range retryQ.Tasks {
		taskQ.Tasks <- retryTask
	}
}

func Schedule(wg *sync.WaitGroup, retryChan chan *tasks.Task) {
	for retryTasks := range retryChan {
		fmt.Println(retryTasks)
	}
}
