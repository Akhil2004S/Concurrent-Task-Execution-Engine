package workers

import (
	"execEngine/tasks"
	"fmt"
	"strings"
	"sync"
)

func Worker(id int, taskQueue *tasks.TaskQueue, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range taskQueue.Tasks {
		success := task.ChangeTaskState(tasks.Running)
		if !success {
			panic("Impossible state transition")
		}
		switch task.TaskType {
		case "add":
			var sum float64
			for _, number := range task.Data {
				if value, dataType := (number.(float64)); dataType {
					sum += value
				} else if value, dataType := (number.(int)); dataType {
					sum += float64(value)
				}
			}
			success = task.ChangeTaskState(tasks.Completed)
			if !success {
				panic("Impossible state transition")
			}

		case "mul":
			product := 1.0
			for _, number := range task.Data {
				if value, dataType := (number.(float64)); dataType {
					product *= value
				} else if value, dataType := (number.(int)); dataType {
					product *= float64(value)
				}
			}
			success = task.ChangeTaskState(tasks.Completed)
			if !success {
				panic("Impossible state transition")
			}

		case "print":
			var messageList []string
			for _, message := range task.Data {
				if value, dataType := (message.(string)); dataType {
					messageList = append(messageList, value)
				}
			}
			finalMessage := strings.Join(messageList, " ")
			fmt.Println(finalMessage)
			success = task.ChangeTaskState(tasks.Completed)
			if !success {
				panic("Impossible state transition")
			}

		default:
			fmt.Println("Invalid operation")
			success = task.ChangeTaskState(tasks.Failed)
			if !success {
				panic("Impossible state transition")
			}
		}

		fmt.Printf("Task %d is executed by worker %d\n", task.Id, id)
		fmt.Printf("Task %d state is %s\n", task.Id, tasks.TaskStateName[task.State])
	}
}
