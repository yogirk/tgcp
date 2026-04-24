package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
)

// Category accents. Each GCP service category gets one colour from the
// banner palette so the home menu, sidebar icons, and category headers
// share a single visual system. Compute/Databases/Data cover the bulk
// of active use; the remaining categories reuse colours deliberately so
// users learn "yellow = runs code, red = holds data, green = ops".
const (
	catCompute  = "Compute"
	catStorage  = "Storage"
	catDatabase = "Databases"
	catData     = "Data & Analytics"
	catSecurity = "Security & Networking"
	catObserv   = "Observability"
	catDevOps   = "DevOps"
	catOverview = "Overview"
)

// categoryColor maps a category name to its accent colour.
var categoryColor = map[string]lipgloss.Color{
	catOverview: styles.ColorBrandAccent,
	catCompute:  styles.ColorAccentYellow,
	catStorage:  styles.ColorAccentRed,
	catDatabase: styles.ColorAccentRed,
	catData:     styles.ColorBrandAccent,
	catSecurity: styles.ColorAccentYellow,
	catObserv:   styles.ColorInfo,
	catDevOps:   styles.ColorAccentGreen,
}

// serviceCategory maps each service short-name to its category. Kept here so
// the sidebar and home menu can look up the colour for a service icon even
// though the sidebar doesn't carry an explicit category header.
var serviceCategory = map[string]string{
	"overview":         catOverview,
	"gce":              catCompute,
	"gke":              catCompute,
	"run":              catCompute,
	"gcs":              catStorage,
	"disks":            catStorage,
	"sql":              catDatabase,
	"spanner":          catDatabase,
	"bigtable":         catDatabase,
	"redis":            catDatabase,
	"firestore":        catDatabase,
	"bq":               catData,
	"dataflow":         catData,
	"dataproc":         catData,
	"pubsub":           catData,
	"iam":              catSecurity,
	"secrets":          catSecurity,
	"net":              catSecurity,
	"logs":             catObserv,
	"cloudbuild":       catDevOps,
	"artifactregistry": catDevOps,
}

// CategoryColor returns the accent colour for a category name. Falls back to
// the subtle border colour for unknown categories.
func CategoryColor(name string) lipgloss.Color {
	if c, ok := categoryColor[name]; ok {
		return c
	}
	return styles.ColorBorderSubtle
}

// ServiceAccent returns the accent colour for a service short-name, based on
// its category. Used to tint service icons.
func ServiceAccent(shortName string) lipgloss.Color {
	if cat, ok := serviceCategory[shortName]; ok {
		return CategoryColor(cat)
	}
	return styles.ColorBrandAccent
}
