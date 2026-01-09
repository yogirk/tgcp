package overview

// -----------------------------------------------------------------------------
// Data Models
// -----------------------------------------------------------------------------

type BillingInfo struct {
	Enabled            bool
	BillingAccountName string
	BillingAccountID   string
}

type Recommendation struct {
	ID                     string
	Description            string
	RecommenderSubtype     string
	Priority               string
	State                  string
	EstimatedSavingsAmount float64
	CurrencyCode           string
}

type ResourceInventory struct {
	InstanceCount int
	DiskCount     int
	DiskGB        int
	IPCount       int
	SQLCount      int
	BucketCount   int
	DatasetCount  int
}

type SpendLimit struct {
	Name            string
	BudgetAmount    string
	CurrencyCode    string
	AlertThresholds []float64
}

type DashboardData struct {
	Info            BillingInfo
	Recommendations []Recommendation
	Inventory       ResourceInventory
	Budgets         []SpendLimit

	// Granular Loading States
	InfoLoading      bool
	RecsLoading      bool
	InventoryLoading bool
	BudgetsLoading   bool

	Error error
}

// Separate Messages for Granular Updates
type InfoMsg BillingInfo
type RecsMsg []Recommendation
type InventoryMsg ResourceInventory
type BudgetsMsg []SpendLimit
