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
