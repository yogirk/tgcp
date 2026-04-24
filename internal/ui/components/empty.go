package components

import (
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
)

// emptyMessages are playful one-liners shown when a resource list returns
// zero rows. Keyed by resource type; the "default" entry is used as a
// fallback. Messages are picked deterministically by hour-of-day so the
// text stays stable during a session but varies day-to-day.
var emptyMessages = map[string][]string{
	"instances": {
		"No instances running. Quiet day in the cloud.",
		"Zero instances. Savings incoming.",
		"No machines here — fresh start.",
	},
	"buckets": {
		"Zero buckets. A clean canvas.",
		"No storage yet — empty shelves.",
	},
	"clusters": {
		"No clusters here. Ready when you are.",
		"Zero clusters — ship's waiting in dock.",
	},
	"databases": {
		"No databases. A fresh slate.",
		"Empty data tier — pristine.",
	},
	"disks": {
		"No disks attached. Light packing.",
	},
	"logs": {
		"The logs are silent. Good news, usually.",
		"Nothing to report — quiet is golden.",
	},
	"recommendations": {
		"No recommendations. You're running a tight ship.",
		"Nothing to optimize — nice.",
	},
	"budgets": {
		"No budgets configured. Set one up to track spend.",
		"Budget radar is clear.",
	},
	"jobs": {
		"No jobs running. Batch queue is empty.",
	},
	"topics": {
		"No topics yet. Ready to publish.",
	},
	"subscriptions": {
		"No subscriptions. Nobody's listening.",
	},
	"secrets": {
		"No secrets stored. Or they're really well hidden.",
	},
	"services": {
		"No services deployed. The stage is yours.",
	},
	"images": {
		"No images in this registry.",
	},
	"builds": {
		"No builds running. All quiet.",
	},
	"default": {
		"Nothing here yet.",
		"Quiet in this corner.",
		"Empty — peaceful, huh?",
	},
}

// EmptyState renders a subtle, italicised one-liner for zero-result views.
// resourceType selects the message pool ("instances", "buckets", ...); an
// unknown key falls back to the "default" pool.
func EmptyState(resourceType string) string {
	pool, ok := emptyMessages[resourceType]
	if !ok || len(pool) == 0 {
		pool = emptyMessages["default"]
	}
	msg := pool[time.Now().Hour()%len(pool)]

	return lipgloss.NewStyle().
		Foreground(styles.ColorTextMuted).
		Italic(true).
		Padding(styles.SpaceS, styles.SpaceM).
		Render("‹ " + msg + " ›")
}
