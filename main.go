package main

import (
	"fmt"
	"sync"

	"execEngine/tasks"
	"execEngine/workers"
)

func main() {
	numWorkers := 4

	taskQueue := &tasks.TaskQueue{}
	taskQueue.Tasks = make(chan *tasks.Task)

	var wg sync.WaitGroup

	for w := range numWorkers {
		wg.Add(1)
		go workers.Worker(w, taskQueue, &wg)
	}

	for id := range 3 {
		err := tasks.CreateTasks(id, "add", taskQueue, 1, 1)
		if err != nil {
			fmt.Println(err)
		}
	}
	close(taskQueue.Tasks)

	wg.Wait()
}
