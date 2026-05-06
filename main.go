package main

import (
	"fmt"
	"sync"

	"execEngine/scheduler"
	"execEngine/tasks"
	"execEngine/workers"
)

func main() {
	numWorkers := 2

	taskQueue := &tasks.TaskQueue{}
	retryQueue := &tasks.RetryQueue{}
	taskQueue.Tasks = make(chan *tasks.Task, 7)
	retryQueue.Tasks = make(chan *tasks.Task, 10)

	var wg sync.WaitGroup
	var taskWg sync.WaitGroup

	// schedWg.Add(1)
	go scheduler.Schedule(taskQueue, retryQueue)

	for w := range numWorkers {
		wg.Add(1)
		go workers.Worker(w, taskQueue, retryQueue, &wg, &taskWg)
	}

	for id := range 3 {
		err := tasks.CreateTasks(id, "add", taskQueue, &taskWg, 1, 1)
		if err != nil {
			fmt.Println(err)
		}
	}

	// for id := range 3 {
	// 	err := tasks.CreateTasks(id, "mul", taskQueue, &taskWg, 1, 1)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// }

	taskWg.Wait()
	close(taskQueue.Tasks)
	close(retryQueue.Tasks)
	wg.Wait()
	// schedWg.Wait()
}
