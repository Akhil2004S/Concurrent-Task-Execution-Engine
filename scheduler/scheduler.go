package scheduler

import (
	"context"
	"execEngine/tasks"
	"fmt"
	"sync"
)

// A simple scheduler that pushes tasks into the queue that the worker consumes
// It handles both waiting queue and retry queue
func Schedule(interrput context.Context, wg *sync.WaitGroup, taskQ *tasks.TaskQueue, retryQ *tasks.RetryQueue, waitQ *tasks.WaitingQueue) {
	fmt.Println("Scheduler active")
	for {
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
			select {
			case taskQ.Tasks <- task:
			case <-interrput.Done():
				wg.Done()
				return
			}

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
			select {
			case taskQ.Tasks <- task:
			case <-interrput.Done():
				wg.Done()
				return
			}
			fmt.Println("-----------------------Task added for retry-----------------------------")
		case <-interrput.Done():
			fmt.Println("Schduler interrupted. Exiting")
			wg.Done()
			return
		}

	}
}
