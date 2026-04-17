package main

import (
	"fmt"
	"sync"
)

type Task struct {
	id       int
	function func(numbers ...int) int
	numTasks int
}

type TaskQueue struct {
	tasks chan Task
}

func CreateTasks(id int, queue *TaskQueue) {
	task := Task{
		id:       id,
		function: add,
	}
	queue.tasks <- task
}

func add(numbers ...int) int {
	sum := 0
	for _, number := range numbers {
		sum += number
	}
	return sum
}

func worker(id int, taskQueue *TaskQueue, wg *sync.WaitGroup, numbers ...int) {
	defer wg.Done()
	for task := range taskQueue.tasks {
		task.function(numbers...)
		fmt.Printf("Task %d is executed by worker %d\n", task.id, id)
	}
}

func main() {
	numWorkers := 10

	taskQueue := &TaskQueue{}
	taskQueue.tasks = make(chan Task)

	var wg sync.WaitGroup

	for w := range numWorkers {
		wg.Add(1)
		go worker(w, taskQueue, &wg, 1, 1)
	}

	for id := range 10 {
		CreateTasks(id, taskQueue)
	}
	close(taskQueue.tasks)

	wg.Wait()
}
