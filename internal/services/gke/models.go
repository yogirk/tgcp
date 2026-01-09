package gke

// -----------------------------------------------------------------------------
// Models
// -----------------------------------------------------------------------------

type Cluster struct {
	Name          string
	Location      string // Zone or Region
	Status        string // RUNNING, RECONCILING, etc.
	MasterVersion string
	Endpoint      string
	Network       string
	Subnetwork    string
	NodeCount     int    // Current node count
	Mode          string // "Standard" or "Autopilot"

	// Full resource name for API calls
	// format: projects/{project}/locations/{location}/clusters/{name}
	SelfLink string

	// Detailed info (loaded on demand or with list if cheap)
	NodePools []NodePool
}

type NodePool struct {
	Name             string
	Status           string // PROVISIONING, RUNNING
	MachineType      string // e.g. e2-standard-4
	DiskSizeGb       int64
	InitialNodeCount int64
	Autoscaling      AutoscalingConfig
	IsSpot           bool
	Version          string
}

type AutoscalingConfig struct {
	Enabled      bool
	MinNodeCount int64
	MaxNodeCount int64
}
