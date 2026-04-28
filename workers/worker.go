package workers

import (
	"context"
	"execEngine/tasks"
	"fmt"
	"strings"
	"sync"
)

type Result struct {
	taskResult  any
	attemptID   int
	isSuccess   bool
	failureData tasks.Failure
}

func Worker(id int, taskQueue *tasks.TaskQueue, retryQueue *tasks.RetryQueue, wg *sync.WaitGroup, taskWg *sync.WaitGroup, schedulerWg *sync.WaitGroup) {
	defer wg.Done()
	defer schedulerWg.Done()
	for task := range taskQueue.Tasks {
		success := task.ChangeTaskState(tasks.Running)
		if !success {
			fmt.Println("================Impossible state transition performed===================")
		}

		ctx, cancel := context.WithTimeout(context.Background(), task.TimeLimit)
		defer cancel()
		resultChan := make(chan Result, 1)
		attemptID := task.RetryData.RetryCount
		go executeTask(ctx, task.Data, task.TaskType, attemptID, resultChan)

		select {
		case result := <-resultChan:
			if attemptID == task.RetryData.RetryCount && result.isSuccess {
				task.ChangeTaskState(tasks.Completed)
				fmt.Printf("Task completed. ID :%d. The result is: %v. Worker Id: %d\n", task.Id, result.taskResult, id)
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

		case <-ctx.Done():
			select {
			case result := <-resultChan:
				fmt.Println("Timed out and completed at the same time. Shit")
				if attemptID == task.RetryData.RetryCount && result.isSuccess {
					task.ChangeTaskState(tasks.Completed)
					fmt.Printf("Task completed. ID :%d. The result is: %v. Worker Id: %d\n", task.Id, result.taskResult, id)
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
			default:
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
		}

		addToQueue := ShouldRetry(task, id)
		fmt.Println("Retry decision:", addToQueue)
		if addToQueue {
			task.RetryData.RetryCount = task.RetryData.RetryCount + 1
			retryQueue.Tasks <- task
		} else {
			taskWg.Done()
		}
	}
}

func executeTask(ctx context.Context, taskData []any, taskType string, attempId int, result chan Result) {
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
			sent = SendResult(ctx, result, computeResult)
		}
		if !sent {
			computeResult := Result{
				taskResult: 0,
				attemptID:  attempId,
				isSuccess:  false,
				failureData: tasks.Failure{
					Type:           "Cancelled",
					Reason:         "Value not received by the channel",
					Classification: tasks.System,
				},
			}
			sent = SendResult(ctx, result, computeResult)
		}
	}()

	switch taskType {
	case "add":
		var sum float64
		for _, number := range taskData {
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
		sent = SendResult(ctx, result, computeResult)

	case "mul":
		product := 1.0
		for _, number := range taskData {
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
		sent = SendResult(ctx, result, computeResult)

	case "print":
		var messageList []string
		for _, message := range taskData {
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
		sent = SendResult(ctx, result, computeResult)

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
		sent = SendResult(ctx, result, computeResult)
	}
}
