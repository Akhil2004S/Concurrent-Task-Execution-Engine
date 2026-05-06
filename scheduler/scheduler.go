package scheduler

import (
	"execEngine/tasks"
	"fmt"
)

func Schedule(taskQ *tasks.TaskQueue, retryQ *tasks.RetryQueue) {
	for retryTasks := range retryQ.Tasks {
		fmt.Println("============================Scheduler Retry=================================")
		fmt.Printf("THE RETRY TASKS. ID: %d\n", retryTasks.Id)
		taskQ.Tasks <- retryTasks
		fmt.Println("-----------------------Task added for retry-----------------------------")
	}
}
