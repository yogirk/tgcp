package cloudsql

import (
	"strings"
	"testing"

	"github.com/yogirk/tgcp/internal/core"
)

func TestCloudSQLView(t *testing.T) {
	cache := core.NewCache()
	svc := NewService(cache)
	svc.projectID = "test-project"

	t.Run("Empty State", func(t *testing.T) {
		svc.instances = nil
		view := svc.View()

		if !strings.Contains(view, "Project test-project") {
			t.Errorf("View missing project ID breadcrumb")
		}
		if !strings.Contains(view, "Cloud SQL") {
			t.Errorf("View missing service name")
		}
		// EmptyState is italicized and includes brackets or markers
		if !strings.Contains(view, "machines") && !strings.Contains(view, "instances") && !strings.Contains(view, "Nothing here yet") && !strings.Contains(view, "Quiet") {
			t.Errorf("View missing empty state message. Got: %q", view)
		}
	})

	t.Run("Populated State", func(t *testing.T) {
		svc.instances = []Instance{
			{Name: "db-1", State: "RUNNABLE", Region: "us-central1"},
		}
		svc.updateTable(svc.instances) // Ensure table has rows
		view := svc.View()

		if !strings.Contains(view, "db-1") {
			t.Errorf("View missing instance name. Got: %q", view)
		}
		if !strings.Contains(view, "1 total") {
			t.Errorf("View missing status summary")
		}
		// StandardTable view might be complex, but it should contain the data
		if !strings.Contains(view, "RUNNABLE") {
			t.Errorf("View missing status")
		}
	})
}
