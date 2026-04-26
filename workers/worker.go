package workers

import (
	"execEngine/tasks"
	"fmt"
	"strings"
	"sync"
	"time"
)

type Result struct {
	taskResult  any
	attemptID   int
	isSuccess   bool
	failureData tasks.Failure
}

func Worker(id int, taskQueue *tasks.TaskQueue, wg *sync.WaitGroup, taskWg *sync.WaitGroup) {
	defer wg.Done()
	var resultChan chan Result
	for task := range taskQueue.Tasks {
		success := task.ChangeTaskState(tasks.Running)
		if !success {
			panic("Impossible state transition performed")
		}

		resultChan = make(chan Result, 1)
		attemptID := task.RetryData.RetryCount
		go executeTask(task, attemptID, resultChan)

		select {
		case result := <-resultChan:
			if attemptID == task.RetryData.RetryCount && result.isSuccess {
				task.ChangeTaskState(tasks.Completed)
				fmt.Printf("Task completed. ID :%d. The result is: %v\n", task.Id, result.taskResult)
			} else if attemptID != task.RetryData.RetryCount && result.isSuccess {
				fmt.Println("The result is ignored due to late completion")
			} else if attemptID != task.RetryData.RetryCount && !result.isSuccess {
				fmt.Println("Ignore. Old execution")
			} else {
				fmt.Println("Task has failed sucessfully")
				task.FailureData = result.failureData
				fmt.Println("Failure is due to", task.FailureData.Reason)
				task.ChangeTaskState(tasks.Failed)
			}

		case _ = <-time.After(task.TimeLimit):
			if attemptID == task.RetryData.RetryCount {
				fmt.Println("Execution timeout")
				task.FailureData = tasks.Failure{
					Type:           "Timeout",
					Reason:         "Exceeded time limit",
					Classification: tasks.Transient,
				}
				task.ChangeTaskState(tasks.Failed)
			} else {
				fmt.Println("Ignored update. Late execution")
			}
		}

		addToQueue := shouldRetry(task, id)
		fmt.Println("Retry decision:", addToQueue)
		if addToQueue {
			task.RetryData.RetryCount = task.RetryData.RetryCount + 1
			time.Sleep(10 * time.Millisecond)
			taskQueue.Tasks <- task
		} else {
			taskWg.Done()
		}
		fmt.Printf("Task: %d Worker: %d State: %s\n", task.Id, id, tasks.TaskStateName[task.State])
		// time.Sleep(1 * time.Millisecond)
	}
}

func executeTask(task *tasks.Task, attempId int, result chan Result) {
	sent := false
	defer func() {
		if r := recover(); r != nil {
			computeResult := Result{
				taskResult: 0,
				attemptID:  attempId,
				isSuccess:  false,
				failureData: tasks.Failure{
					Type:           "Panic",
					Reason:         r,
					Classification: tasks.System,
				},
			}
			result <- computeResult
			sent = true
			return
		}
		if !sent {
			computeResult := Result{
				taskResult: 0,
				attemptID:  attempId,
				isSuccess:  false,
				failureData: tasks.Failure{
					Type:           "Not sent",
					Reason:         "Value not received by the channel",
					Classification: tasks.System,
				},
			}
			result <- computeResult
			sent = true
		}
	}()

	switch task.TaskType {
	case "add":
		var sum float64
		for _, number := range task.Data {
			if value, ok := (number.(float64)); ok {
				sum += value
			} else if value, ok := (number.(int)); ok {
				sum += float64(value)
			}
		}
		computeResult := Result{
			taskResult:  sum,
			attemptID:   attempId,
			isSuccess:   true,
			failureData: tasks.Failure{},
		}
		result <- computeResult

	case "mul":
		product := 1.0
		for _, number := range task.Data {
			if value, dataType := (number.(float64)); dataType {
				product *= value
			} else if value, dataType := (number.(int)); dataType {
				product *= float64(value)
			}
		}
		computeResult := Result{
			taskResult:  product,
			attemptID:   attempId,
			isSuccess:   true,
			failureData: tasks.Failure{},
		}
		result <- computeResult

	case "print":
		var messageList []string
		for _, message := range task.Data {
			if value, dataType := (message.(string)); dataType {
				messageList = append(messageList, value)
			}
		}
		finalMessage := strings.Join(messageList, " ")
		fmt.Println(finalMessage)
		computeResult := Result{
			taskResult:  finalMessage,
			attemptID:   attempId,
			isSuccess:   true,
			failureData: tasks.Failure{},
		}
		result <- computeResult

	default:
		computeResult := Result{
			taskResult: 0,
			attemptID:  attempId,
			isSuccess:  false,
			failureData: tasks.Failure{
				Type:           "User Error",
				Reason:         "Invalid input",
				Classification: tasks.Permanent,
			},
		}
		result <- computeResult
	}
}

func shouldRetry(task *tasks.Task, workerId int) bool {
	if task.State == tasks.Failed {
		fmt.Printf("Task %d (%s). State: %s Retry Count: %d. Error Type:%s. Worker %d\n", task.Id, task.TaskType, tasks.TaskStateName[task.State], task.RetryData.RetryCount, tasks.ClassificationName[task.FailureData.Classification], workerId)
		if task.FailureData.Classification != tasks.Permanent && task.RetryData.RetryCount < task.RetryData.RetryLimit {
			return true
		}
		return false
	}
	return false
}
