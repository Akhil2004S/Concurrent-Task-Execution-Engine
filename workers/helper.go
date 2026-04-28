package workers

import (
	"context"
	"execEngine/tasks"
	"fmt"
)

func SendResult(ctx context.Context, ch chan Result, result Result) bool {
	select {
	case ch <- result:
		return true
	case <-ctx.Done():
		return false
	}
}

func ShouldRetry(task *tasks.Task, workerId int) bool {
	if task.State == tasks.Failed {
		fmt.Printf("Task %d (%s). State: %s Retry Count: %d. Error Type:%s. Worker %d\n", task.Id, task.TaskType, tasks.TaskStateName[task.State], task.RetryData.RetryCount, tasks.ClassificationName[task.FailureData.Classification], workerId)
		if task.FailureData.Classification != tasks.Permanent && task.RetryData.RetryCount < task.RetryData.RetryLimit {
			return true
		}
		return false
	}
	return false
}
