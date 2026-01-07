package cloudsql

// InstanceState represents the status of a Cloud SQL instance
type InstanceState string

const (
	StateRunnable    InstanceState = "RUNNABLE"
	StateSuspended   InstanceState = "SUSPENDED"
	StatePending     InstanceState = "PENDING_CREATE"
	StateMaintenance InstanceState = "MAINTENANCE"
	StateFailed      InstanceState = "FAILED"
	StateUnknown     InstanceState = "UNKNOWN"
)

// Instance represents a simplified Cloud SQL instance
type Instance struct {
	Name            string
	ProjectID       string
	Region          string
	DatabaseVersion string
	State           InstanceState
	Tier            string
	PrimaryIP       string
	ConnectionName  string

	// Details
	StorageGB  int64
	AutoBackup bool
	Activation string // ALWAYS or NEVER
}
