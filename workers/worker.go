package workers

import (
	"execEngine/tasks"
	"fmt"
	"strings"
	"sync"
)

func Worker(id int, taskQueue *tasks.TaskQueue, wg *sync.WaitGroup, taskWg *sync.WaitGroup) {
	defer wg.Done()
	for task := range taskQueue.Tasks {
		executeTask(task)
		addToQueue := shouldRetry(task)
		fmt.Println("Retry decision:", addToQueue)
		if addToQueue {
			task.RetryData.RetryCount = task.RetryData.RetryCount + 1
			taskQueue.Tasks <- task
		} else {
			taskWg.Done()
		}
		fmt.Printf("Task: %d Worker: %d State: %s\n", task.Id, id, tasks.TaskStateName[task.State])
		// time.Sleep(1 * time.Millisecond)
	}
}

func executeTask(task *tasks.Task) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("panic occurred")
			task.FailureData = tasks.Failure{
				Type:           "Panic",
				Reason:         r,
				Classification: tasks.System,
			}
			task.ChangeTaskState(tasks.Failed)
		}
	}()
	success := task.ChangeTaskState(tasks.Running)
	if !success {
		panic("Impossible state transition")
	}
	switch task.TaskType {
	case "add":
		success := task.ChangeTaskState(tasks.Running)
		if !success {
			panic("Impossible state transition")
		}
		fmt.Println("add")
		var sum float64
		for _, number := range task.Data {
			if value, ok := (number.(float64)); ok {
				sum += value
			} else if value, ok := (number.(int)); ok {
				sum += float64(value)
			}
		}
		success = task.ChangeTaskState(tasks.Completed)
		if !success {
			panic("Impossible state transition")
		}

	case "mul":
		product := 1.0
		for _, number := range task.Data {
			if value, dataType := (number.(float64)); dataType {
				product *= value
			} else if value, dataType := (number.(int)); dataType {
				product *= float64(value)
			}
		}
		success = task.ChangeTaskState(tasks.Completed)
		if !success {
			panic("Impossible state transition")
		}

	case "print":
		var messageList []string
		for _, message := range task.Data {
			if value, dataType := (message.(string)); dataType {
				messageList = append(messageList, value)
			}
		}
		finalMessage := strings.Join(messageList, " ")
		fmt.Println(finalMessage)
		success = task.ChangeTaskState(tasks.Completed)
		if !success {
			panic("Impossible state transition")
		}

	default:
		fmt.Println("Invalid operation")
		task.FailureData = tasks.Failure{
			Type:           "User Error",
			Reason:         "Invalid input",
			Classification: tasks.Permanent,
		}
		success := task.ChangeTaskState(tasks.Failed)
		if !success {
			panic("Impossible state transition")
		}
	}
}

func shouldRetry(task *tasks.Task) bool {
	if task.State == tasks.Failed {
		fmt.Printf("Task %d of type %s. State: %s Retry Count: %d\n", task.Id, tasks.TaskStateName[task.State], tasks.ClassificationName[task.FailureData.Classification], task.RetryData.RetryCount)
		if task.FailureData.Classification != tasks.Permanent && task.RetryData.RetryCount < task.RetryData.RetryLimit {
			return true
		}
		return false
	}
	return false
}
