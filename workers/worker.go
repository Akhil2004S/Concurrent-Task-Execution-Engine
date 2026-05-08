package workers

import (
	"context"
	"execEngine/tasks"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Struct to store the result of the task completed
type Result struct {
	taskResult  any
	attemptID   int
	isSuccess   bool
	failureData tasks.Failure
}

// Worker that picks a task from the queue and executes it
func Worker(interrupt context.Context, id int, taskQueue *tasks.TaskQueue, retryQueue *tasks.RetryQueue, waitingQueue *tasks.WaitingQueue, metrics *tasks.Metrics, wg *sync.WaitGroup, taskWg *sync.WaitGroup) {

	defer wg.Done()
	for {
		select {
		case <-interrupt.Done():
			fmt.Println("--------------------------INTERRUPT-------------------------------------. Worker ID:", id)
			return

		case task, ok := <-taskQueue.Tasks:
			if !ok {
				return
			}
			task.ExecutionStartedAt = time.Now() // Storing the time at which execution started
			// Task state updated before starting execution
			success := task.ChangeTaskState(tasks.Running)
			if !success {
				fmt.Println("================Impossible state transition performed===================")
			}

			// Context for the implementing timeout of tasks
			ctx, cancel := context.WithTimeout(context.Background(), task.TimeLimit)

			// Channel to store the result of the task. It is of the result struct type
			resultChan := make(chan Result, 1)

			// The attempt ID is the attempt number of the task. Stored to check for stale updates
			attemptID := task.RetryData.RetryCount
			go executeTask(ctx, task.Data, task.TaskType, attemptID, resultChan)

			// Blocking operation that waits for either the result or a timeout
			// The result can be a success or a failure
			select {
			case result := <-resultChan:
				computeResult(id, task, &result, attemptID, metrics)
			// Handles task timeout
			case <-ctx.Done():
				select {
				// Checks if the result exists when a timeout occurs meaning that both the execution completion and timeout fired at the same timing
				case result := <-resultChan:
					computeResult(id, task, &result, attemptID, metrics)
				// If the task is timed out without any result, this block updates the necessary stuff
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

			// Checks if failed tasks should retry based on the classification of the failure
			addToQueue := ShouldRetry(task, id, metrics)
			// fmt.Println("Retry decision:", addToQueue)
			if addToQueue {
				// fmt.Printf("The following task is gonna be retried. Task id:%d, reason:%s\n", task.Id, task.FailureData.Reason)
				task.RetryData.RetryCount++
				atomic.AddInt64(&metrics.RetryCount, 1)
				select {
				case retryQueue.Tasks <- task:
				case <-interrupt.Done():
					taskWg.Done()
				}
			} else {
				taskWg.Done()
			}
			cancel()
		}
	}
}

func computeResult(workerId int, task *tasks.Task, result *Result, attemptID int, metrics *tasks.Metrics) {
	if attemptID == task.RetryData.RetryCount && result.isSuccess {
		task.ChangeTaskState(tasks.Completed) // State change of task
		atomic.AddInt64(&metrics.TotalLatencyMs, int64(time.Since(task.SubmittedAt)))
		atomic.AddInt64(&metrics.TotalExecutionLatencyMs, int64(time.Since(task.ExecutionStartedAt)))
		atomic.AddInt64(&metrics.CompletedTasks, 1)
		fmt.Printf("Task completed. ID: %d. The result is: %v. Worker Id: %d\n", task.Id, result.taskResult, workerId)
	} else if attemptID != task.RetryData.RetryCount && result.isSuccess {
		// This is stale update where old goroutine execution tries to update the state
		fmt.Println("The result is ignored due to late completion")
	} else if attemptID != task.RetryData.RetryCount && !result.isSuccess {
		fmt.Println("Ignore. Old execution failed")
	} else {
		fmt.Println("Task has failed sucessfully")
		task.FailureData = result.failureData
		fmt.Println("Failure is due to", task.FailureData.Reason)
		task.ChangeTaskState(tasks.Failed)
	}
}

// The execution logic of the tasks
// Here results are sent to the result channel as a struct which is caught by the worker
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
		} else if !sent {
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
