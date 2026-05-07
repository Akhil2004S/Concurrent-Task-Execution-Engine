package scheduler

import (
	"execEngine/tasks"
	"fmt"
	"sync"
)

// A simple scheduler that pushes tasks into the queue that the worker consumes
// It handles both waiting queue and retry queue
func Schedule(wg *sync.WaitGroup, taskQ *tasks.TaskQueue, retryQ *tasks.RetryQueue, waitQ *tasks.WaitingQueue) {
	fmt.Println("Scheduler active")
	for {
		fmt.Printf("Len of task queue:%d, capacity of task queue:%d\n", len(taskQ.Tasks), cap(taskQ.Tasks))
		select {
		case task, ok := <-waitQ.Tasks:
			if !ok {
				waitQ.Tasks = nil
				if waitQ.Tasks == nil && retryQ.Tasks == nil {
					fmt.Println("Channels closed. Scehduler exiting...")
					wg.Done()
					return
				}
				continue
			}
			// Add waiting task to main task queue
			taskQ.Tasks <- task
		case task, ok := <-retryQ.Tasks:
			if !ok {
				retryQ.Tasks = nil
				if waitQ.Tasks == nil && retryQ.Tasks == nil {
					fmt.Println("Channels closed. Scehduler exiting...")
					wg.Done()
					return
				}
				continue
			}
			// Add retried task to main task queue
			fmt.Println("============================Scheduler Retry=================================")
			fmt.Printf("THE RETRY TASKS. ID: %d\n", task.Id)
			taskQ.Tasks <- task
			fmt.Println("-----------------------Task added for retry-----------------------------")
		}

	}
}
