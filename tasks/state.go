package tasks

type TaskState int

// All possible state for a task
const (
	Pending TaskState = iota
	Running
	Completed
	Failed
)

// Names for task for metrics
var TaskStateName = map[TaskState]string{
	Pending:   "pending",
	Running:   "running",
	Completed: "completed",
	Failed:    "failed",
}

// Logic to change the state of the task. Terminal state being completed and failed.
func (task *Task) ChangeTaskState(newState TaskState) bool {
	// Returns if true if the transition is a success
	if (task.State == Pending && newState == Running) || (task.State == Running && newState == Completed) || (task.State == Running && newState == Failed) || (task.State == Failed && newState == Running) {
		task.State = newState
		return true
	} else {
		return false
	}
}
