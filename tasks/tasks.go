package tasks

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// Task info struct
type Task struct {
	Id                 int
	TaskType           string
	Data               []any
	State              TaskState
	FailureData        Failure
	RetryData          Retry
	TimeLimit          time.Duration
	SubmittedAt        time.Time
	ExecutionStartedAt time.Time
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
func CreateTasks(id int, taskType string, waitQ *WaitingQueue, metrics *Metrics, taskWg *sync.WaitGroup, data ...any) error {
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
		TimeLimit:   10 * time.Millisecond,
		SubmittedAt: time.Now(),
	}
	select {
	case waitQ.Tasks <- task:
		taskWg.Add(1)
		atomic.AddInt64(&metrics.TotalTasks, 1)
		// fmt.Printf("Length of queue:%d and capacity of queue:%d\n", len(waitQ.Tasks), cap(waitQ.Tasks))
	default:
		atomic.AddInt64(&metrics.DroppedTasks, 1)
		atomic.AddInt64(&metrics.TotalTasks, 1)
	}
	return nil
}
