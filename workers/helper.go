package workers

import (
	"context"
	"execEngine/tasks"
	"fmt"
	"sync/atomic"
)

// Sends the result to the channel
func SendResult(ctx context.Context, ch chan Result, result Result) bool {
	select {
	case ch <- result:
		return true
	case <-ctx.Done():
		return false
	}
}

// Logic to check whether a task should be retried. Returns true if the task should retry
func ShouldRetry(task *tasks.Task, workerId int, metrics *tasks.Metrics) bool {
	if task.State == tasks.Failed {
		fmt.Printf("Task %d (%s). State: %s Retry Count: %d. Error Type:%s. Worker %d\n", task.Id, task.TaskType, tasks.TaskStateName[task.State], task.RetryData.RetryCount, tasks.ClassificationName[task.FailureData.Classification], workerId)
		if task.FailureData.Classification != tasks.Permanent && task.RetryData.RetryCount < task.RetryData.RetryLimit {
			return true
		}
		atomic.AddInt64(&metrics.FailedTasks, 1)
		return false
	}
	return false
}
