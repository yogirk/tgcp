package dataflow

type Job struct {
	ID             string
	Name           string
	Type           string // JOB_TYPE_STREAMING / BATCH
	State          string // JOB_STATE_RUNNING
	CreateTime     string
	Location       string
	CurrentWorkers int64 // Derived if available, or just from metric
}
