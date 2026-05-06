package tasks

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Task info struct
type Task struct {
	Id          int
	TaskType    string
	Data        []any
	State       TaskState
	FailureData Failure
	RetryData   Retry
	TimeLimit   time.Duration
}

// Queue of tasks that the worker should pick and execute
type TaskQueue struct {
	Tasks chan *Task
}

// Queue of tasks that should be retried. Put back to the main queue by scheduler
type RetryQueue struct {
	Tasks chan *Task
}

// Queue of tasks that are waiting. These tasks will be pushed to the main queue by the scheduler
type WaitingQueue struct {
	Tasks chan *Task
}

// Create task which increments the wg counter of tasks and pushes the task to the task queue
func CreateTasks(id int, taskType string, waitQ *WaitingQueue, taskWg *sync.WaitGroup, data ...any) error {
	switch taskType {
	case "add":
		for _, value := range data {
			switch value.(type) {
			case int:
			case float64:
			default:
				return errors.New("Invalid data. Requires numbers (int or float)")
			}
		}

	case "mul":
		for _, value := range data {
			switch value.(type) {
			case int:
			case float64:
			default:
				return errors.New("Invalid data. Requires numbers (int or float)")
			}
		}

	case "print":
		for _, value := range data {
			switch value.(type) {
			case string:
			default:
				return errors.New("Invalid data. Requires string")
			}
		}
	}
	task := &Task{
		Id:          id,
		TaskType:    taskType,
		Data:        data,
		State:       Pending,
		FailureData: Failure{},
		RetryData:   Retry{RetryLimit: 3},
		TimeLimit:   1 * time.Nanosecond,
	}
	select {
	case waitQ.Tasks <- task:
		taskWg.Add(1)
		fmt.Println("Task added to waiting queue")
	default:
		fmt.Println("Task NOT added to queue (size full). Try later")
	}
	return nil
}
