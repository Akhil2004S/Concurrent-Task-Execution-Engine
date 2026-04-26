package main

import (
	"fmt"
	"sync"

	"execEngine/tasks"
	"execEngine/workers"
)

func main() {
	numWorkers := 2

	taskQueue := &tasks.TaskQueue{}
	taskQueue.Tasks = make(chan *tasks.Task)

	var wg sync.WaitGroup
	var taskWg sync.WaitGroup

	for w := range numWorkers {
		wg.Add(1)
		go workers.Worker(w, taskQueue, &wg, &taskWg)
	}

	for id := range 3 {
		err := tasks.CreateTasks(id, "add", taskQueue, &taskWg, 1, 1)
		if err != nil {
			fmt.Println(err)
		}
	}

	// for id := range 1 {
	// 	err := tasks.CreateTasks(id, "mul", taskQueue, &taskWg, 1, 1)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// }

	taskWg.Wait()
	close(taskQueue.Tasks)
	wg.Wait()
}
