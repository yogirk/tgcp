package spanner

type Instance struct {
	Name            string // Short ID
	DisplayName     string
	ProjectID       string
	Config          string // regional-us-central1
	State           string // READY
	NodeCount       int
	ProcessingUnits int
	Labels          map[string]string
}
