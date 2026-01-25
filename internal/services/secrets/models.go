package secrets

import "time"

// Secret represents a Secret Manager secret
type Secret struct {
	Name        string            // Short name (e.g., "api-key")
	FullName    string            // Full resource name (projects/xxx/secrets/yyy)
	CreateTime  time.Time
	Labels      map[string]string
	Replication string            // "automatic" or region list
	VersionCount int              // Number of versions
}

// SecretVersion represents a version of a secret
type SecretVersion struct {
	Name       string    // Version number (e.g., "1", "2", "latest")
	FullName   string    // Full resource name
	State      string    // ENABLED, DISABLED, DESTROYED
	CreateTime time.Time
}
