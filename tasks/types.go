package tasks

// File containing all the failure names and stuff. It is self explanatory

type FailureClassification int

const (
	Permanent FailureClassification = iota
	Transient
	System
)

var ClassificationName = map[FailureClassification]string{
	Permanent: "Permanent",
	Transient: "Transient",
	System:    "System",
}

type Failure struct {
	Type           string
	Reason         any
	Classification FailureClassification
}

type Retry struct {
	RetryCount int
	RetryLimit int
}

type Metrics struct {
	TotalTasks              int64   // Create Tasks
	CompletedTasks          int64   // Worker
	DroppedTasks            int64   // Worker
	FailedTasks             int64   // Worker
	RetryCount              int64   // Worker
	TotalLatencyMs          int64   // Worker
	TotalExecutionLatencyMs int64   // Worker
	AvgLatency              int64   // Main
	AvgExecutionLatency     int64   // Main
	RetryRate               float64 // Main
	DropRate                float64 // Main
	FailureRate             float64 // Main
}
