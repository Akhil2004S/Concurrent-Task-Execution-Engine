package main

import (
	"fmt"
	"sync"

	"execEngine/scheduler"
	"execEngine/tasks"
	"execEngine/workers"
)

func main() {
	// Set the number of workers manually. Should be based on the resources available
	numWorkers := 2

	taskQueue := &tasks.TaskQueue{}
	retryQueue := &tasks.RetryQueue{}
	waitingQueue := &tasks.WaitingQueue{}
	taskQueue.Tasks = make(chan *tasks.Task, 3)
	retryQueue.Tasks = make(chan *tasks.Task, 1)
	waitingQueue.Tasks = make(chan *tasks.Task, 2)

	var wg sync.WaitGroup
	var taskWg sync.WaitGroup
	var schedulerWg sync.WaitGroup

	schedulerWg.Add(1)
	go scheduler.Schedule(&schedulerWg, taskQueue, retryQueue, waitingQueue)

	for w := range numWorkers {
		wg.Add(1)
		go workers.Worker(w, taskQueue, retryQueue, &wg, &taskWg)
	}

	// Creating dummy tasks that add two numbers for testing purposes
	for id := range 3 {
		err := tasks.CreateTasks(id, "add", waitingQueue, &taskWg, 1, 1)
		if err != nil {
			fmt.Println(err)
		}
	}

	// Makes sure that the queues are not closed until all the tasks are completed
	taskWg.Wait()
	close(waitingQueue.Tasks)
	close(retryQueue.Tasks)
	schedulerWg.Wait()
	close(taskQueue.Tasks)
	wg.Wait()
}
