package tasks

import (
	"errors"
	"sync"
	"time"
)

type Task struct {
	Id          int
	TaskType    string
	Data        []any
	State       TaskState
	FailureData Failure
	RetryData   Retry
	TimeLimit   time.Duration
}

type TaskQueue struct {
	Tasks chan *Task
}

type RetryQueue struct {
	Tasks chan *Task
}

func CreateTasks(id int, taskType string, queue *TaskQueue, taskWg *sync.WaitGroup, data ...any) error {
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
		RetryData:   Retry{RetryLimit: 1},
		TimeLimit:   1 * time.Nanosecond,
	}
	queue.Tasks <- task
	taskWg.Add(1)
	return nil
}
