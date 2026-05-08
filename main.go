package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"execEngine/scheduler"
	"execEngine/tasks"
	"execEngine/workers"
)

func main() {
	// Set the number of workers manually. Should be based on the resources available
	interrupt, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT)
	defer stop()

	// numWorkers := runtime.NumCPU()
	numWorkers := 8

	taskQueue := &tasks.TaskQueue{}
	retryQueue := &tasks.RetryQueue{}
	waitingQueue := &tasks.WaitingQueue{}
	taskQueue.Tasks = make(chan *tasks.Task, 10000)
	retryQueue.Tasks = make(chan *tasks.Task, 5000)
	waitingQueue.Tasks = make(chan *tasks.Task, 50000)

	taskMetric := &tasks.Metrics{}

	var wg sync.WaitGroup
	var taskWg sync.WaitGroup
	var schedulerWg sync.WaitGroup

	schedulerWg.Add(1)
	go scheduler.Schedule(interrupt, &schedulerWg, taskQueue, retryQueue, waitingQueue)

	// w in the loop will be the id of the worker
	for w := range numWorkers {
		wg.Add(1)
		go workers.Worker(interrupt, w, taskQueue, retryQueue, waitingQueue, taskMetric, &wg, &taskWg)
	}

	// Creating dummy tasks that add two numbers for testing purposes
	for id := range 100000 {
		err := tasks.CreateTasks(id, "add", waitingQueue, taskMetric, &taskWg, 1, 1)
		if err != nil {
			fmt.Println(err)
		}
	}

	// for id := range 100000 {
	// 	err := tasks.CreateTasks(id, "mul", waitingQueue, taskMetric, &taskWg, 1, 1)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// }

	waitDone := make(chan struct{})
	go func() {
		taskWg.Wait()
		select {
		case <-interrupt.Done():
			return
		default:
			close(waitDone)
		}
	}()

	select {
	case <-waitDone:
		shutdown(interrupt, waitingQueue, retryQueue, taskQueue, &schedulerWg, &wg)
	case <-interrupt.Done():
		shutdown(interrupt, waitingQueue, retryQueue, taskQueue, &schedulerWg, &wg)
	}
	// Makes sure that the queues are not closed until all the tasks are completed

	taskMetric.AvgLatency = taskMetric.TotalLatencyMs / (taskMetric.CompletedTasks)
	taskMetric.AvgExecutionLatency = taskMetric.TotalExecutionLatencyMs / (taskMetric.CompletedTasks)
	taskMetric.RetryRate = (float64(taskMetric.RetryCount) / float64(taskMetric.TotalTasks)) * 100
	taskMetric.FailureRate = (float64(taskMetric.FailedTasks) / float64(taskMetric.TotalTasks)) * 100
	taskMetric.DropRate = (float64(taskMetric.DroppedTasks) / float64(taskMetric.TotalTasks)) * 100

	fmt.Println("========================METRICS========================")
	fmt.Printf("Total Tasks: %d\nCompleted Tasks:%d\nDropped Tasks:%d\nDrop Rate: %.2f%%\nRetry Count:%d\nRetry rate:%.2f%%\nFailed Tasks: %d\nFailure Rate: %.2f%%\nTotal Exec Latency(ms):%v\nAverage Exec latency(ms):%v\nTotal Latency:%v\nAverage Latency:%v\n", taskMetric.TotalTasks, taskMetric.CompletedTasks, taskMetric.DroppedTasks, taskMetric.DropRate, taskMetric.RetryCount, taskMetric.RetryRate, taskMetric.FailedTasks, taskMetric.FailureRate, time.Duration(taskMetric.TotalExecutionLatencyMs), time.Duration(taskMetric.AvgExecutionLatency), time.Duration(taskMetric.TotalLatencyMs), time.Duration(taskMetric.AvgLatency))
}

func shutdown(interrupt context.Context, waitQ *tasks.WaitingQueue, retryQ *tasks.RetryQueue, taskQ *tasks.TaskQueue, schedulerWg *sync.WaitGroup, workerWg *sync.WaitGroup) {
	close(waitQ.Tasks)
	select {
	case <-interrupt.Done():
		// interrupted - don't close retryQ, workers might still write to it
	default:
		close(retryQ.Tasks)
	}
	schedulerWg.Wait()
	close(taskQ.Tasks)
	workerWg.Wait()
}
