package tasks

type TaskState int

const (
	Pending TaskState = iota
	Running
	Completed
	Failed
)

var TaskStateName = map[TaskState]string{
	Pending:   "pending",
	Running:   "running",
	Completed: "completed",
	Failed:    "failed",
}

func (task *Task) ChangeTaskState(newState TaskState) bool {
	// Returns if true if the transition is a success
	if (task.State == Pending && newState == Running) || (task.State == Running && newState == Completed) || (task.State == Running && newState == Failed) || (task.State == Failed && newState == Running) {
		task.State = newState
		return true
	} else {
		return false
	}
}
