package iam

// ServiceAccount represents a Google Cloud Service Account
type ServiceAccount struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Disabled    bool   `json:"disabled"`
	UniqueID    string `json:"uniqueId"`
}

// PolicyMember represents a member in an IAM policy
type PolicyMember struct {
	Role   string
	Member string // user:email, serviceAccount:email, etc
}
