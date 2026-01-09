package bigtable

type Instance struct {
	Name        string // Short ID
	DisplayName string
	ProjectID   string
	State       string // READY
	Type        string // PRODUCTION / DEVELOPMENT
}

type Cluster struct {
	Name        string
	Zone        string
	ServeNodes  int
	State       string
	StorageType string // SSD / HDD
}
