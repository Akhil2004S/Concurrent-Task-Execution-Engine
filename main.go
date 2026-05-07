package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"execEngine/scheduler"
	"execEngine/tasks"
	"execEngine/workers"
)

func main() {
	// Set the number of workers manually. Should be based on the resources available
	numWorkers := runtime.NumCPU()

	taskQueue := &tasks.TaskQueue{}
	retryQueue := &tasks.RetryQueue{}
	waitingQueue := &tasks.WaitingQueue{}
	taskQueue.Tasks = make(chan *tasks.Task, 100)
	retryQueue.Tasks = make(chan *tasks.Task, 50)
	waitingQueue.Tasks = make(chan *tasks.Task, 750)

	taskMetric := &tasks.Metrics{}

	var wg sync.WaitGroup
	var taskWg sync.WaitGroup
	var schedulerWg sync.WaitGroup

	schedulerWg.Add(1)
	go scheduler.Schedule(&schedulerWg, taskQueue, retryQueue, waitingQueue)

	for w := range numWorkers {
		wg.Add(1)
		go workers.Worker(w, taskQueue, retryQueue, taskMetric, &wg, &taskWg)
	}

	// Creating dummy tasks that add two numbers for testing purposes
	for id := range 1000 {
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

	// Makes sure that the queues are not closed until all the tasks are completed
	taskWg.Wait()
	close(waitingQueue.Tasks)
	close(retryQueue.Tasks)
	schedulerWg.Wait()
	close(taskQueue.Tasks)
	wg.Wait()

	taskMetric.AvgLatency = taskMetric.TotalLatencyMs / (taskMetric.CompletedTasks)
	taskMetric.AvgExecutionLatency = taskMetric.TotalExecutionLatencyMs / (taskMetric.CompletedTasks)
	taskMetric.RetryRate = (float64(taskMetric.RetryCount) / float64(taskMetric.TotalTasks)) * 100
	taskMetric.FailureRate = (float64(taskMetric.FailedTasks) / float64(taskMetric.TotalTasks)) * 100
	taskMetric.DropRate = (float64(taskMetric.DroppedTasks) / float64(taskMetric.TotalTasks)) * 100

	fmt.Println("========================METRICS========================")
	fmt.Printf("Total Tasks: %d\nCompleted Tasks:%d\nDropped Tasks:%d\nDrop Rate: %.2f%%\nRetry Count:%d\nRetry rate:%.2f%%\nFailed Tasks: %d\nFailure Rate: %.2f%%\nTotal Exec Latency(ms):%v\nAverage Exec latency(ms):%v\nTotal Latency:%v\nAverage Latency:%v\n", taskMetric.TotalTasks, taskMetric.CompletedTasks, taskMetric.DroppedTasks, taskMetric.DropRate, taskMetric.RetryCount, taskMetric.RetryRate, taskMetric.FailedTasks, taskMetric.FailureRate, time.Duration(taskMetric.TotalExecutionLatencyMs), time.Duration(taskMetric.AvgExecutionLatency), time.Duration(taskMetric.TotalLatencyMs), time.Duration(taskMetric.AvgLatency))
}
