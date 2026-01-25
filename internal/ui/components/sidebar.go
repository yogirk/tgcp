package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yogirk/tgcp/internal/styles"
)

type ServiceItem struct {
	Name      string
	ShortName string
	Icon      string
	Active    bool
	IsComing  bool
}

type SidebarModel struct {
	Items   []ServiceItem
	Cursor  int
	Active  bool // Is sidebar focused?
	Visible bool
	Width   int
	Height  int
}

// groupBreaks defines indices after which a visual gap appears (0-indexed)
// This creates subtle spacing between service categories
var groupBreaks = map[int]bool{
	0:  true, // After Overview
	3:  true, // After Compute (GCE, GKE, Cloud Run)
	5:  true, // After Storage (GCS, Disks)
	10: true, // After Databases (Cloud SQL, Spanner, Bigtable, Memorystore, Firestore)
	14: true, // After Data & Analytics (BigQuery, Dataflow, Dataproc, Pub/Sub)
	17: true, // After Security & Networking (IAM, Secrets, Networking)
	18: true, // After Observability (Cloud Logging)
}

func NewSidebar() SidebarModel {
	return SidebarModel{
		Items: []ServiceItem{
			// Overview (top-level)
			{Name: "Overview", ShortName: "overview", Icon: "◉", Active: true},
			// Compute
			{Name: "Compute Engine", ShortName: "gce", Icon: "⚙"},
			{Name: "Kubernetes", ShortName: "gke", Icon: "☸"},
			{Name: "Cloud Run", ShortName: "run", Icon: "▷"},
			// Storage
			{Name: "Cloud Storage", ShortName: "gcs", Icon: "▤"},
			{Name: "Disks", ShortName: "disks", Icon: "◔"},
			// Databases
			{Name: "Cloud SQL", ShortName: "sql", Icon: "⛁"},
			{Name: "Spanner", ShortName: "spanner", Icon: "⬡"},
			{Name: "Bigtable", ShortName: "bigtable", Icon: "▦"},
			{Name: "Memorystore", ShortName: "redis", Icon: "◇"},
			{Name: "Firestore", ShortName: "firestore", Icon: "◲"},
			// Data & Analytics
			{Name: "BigQuery", ShortName: "bq", Icon: "⊞"},
			{Name: "Dataflow", ShortName: "dataflow", Icon: "⇢"},
			{Name: "Dataproc", ShortName: "dataproc", Icon: "⎈"},
			{Name: "Pub/Sub", ShortName: "pubsub", Icon: "⇌"},
			// Security & Networking
			{Name: "IAM", ShortName: "iam", Icon: "⚿"},
			{Name: "Secrets", ShortName: "secrets", Icon: "✦"},
			{Name: "Networking", ShortName: "net", Icon: "⇄"},
			// Observability
			{Name: "Cloud Logging", ShortName: "logs", Icon: "☰"},
		},
		Cursor:  0,
		Active:  true, // Default focus on start
		Visible: true,
		Width:   25,
	}
}

func (m SidebarModel) Init() tea.Cmd {
	return nil
}

func (m SidebarModel) Update(msg tea.Msg) (SidebarModel, tea.Cmd) {
	if !m.Active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Items)-1 {
				m.Cursor++
			}
			// 'enter' usually selects the service to load in main view
			// This will be handled by parent model inspecting the selection
		}
	}
	return m, nil
}

func (m SidebarModel) View() string {
	if !m.Visible {
		return ""
	}

	doc := strings.Builder{}

	// Header
	doc.WriteString(styles.HeaderStyle.Render("SERVICES"))
	doc.WriteString("\n\n")

	for i, item := range m.Items {
		// Format: "icon name" with consistent spacing
		displayName := item.Icon + " " + item.Name
		if item.IsComing {
			displayName += " *"
		}

		isSelected := m.Cursor == i

		var renderedItem string
		if isSelected {
			if m.Active {
				renderedItem = styles.SelectedItemStyle.Render(displayName)
			} else {
				// Selected but not focused (dimmed)
				renderedItem = styles.UnselectedItemStyle.Copy().Foreground(styles.ColorBrandAccent).Render(displayName)
			}
		} else {
			style := styles.UnselectedItemStyle
			if item.IsComing {
				style = style.Copy().Foreground(styles.ColorTextMuted)
			}
			renderedItem = style.Render(displayName)
		}

		doc.WriteString(renderedItem)
		doc.WriteString("\n")

		// Add subtle spacing after group breaks
		if groupBreaks[i] {
			doc.WriteString("\n")
		}
	}

	// Fill remaining height with empty space to maintain border
	// This ensures the sidebar vertical line goes all the way down
	lines := strings.Count(doc.String(), "\n")
	if m.Height > lines {
		doc.WriteString(strings.Repeat("\n", m.Height-lines))
	}

	return styles.SidebarStyle.Width(m.Width).Height(m.Height).Render(doc.String())
}

// Helper to get selected service
func (m SidebarModel) SelectedService() ServiceItem {
	if len(m.Items) == 0 {
		return ServiceItem{}
	}
	return m.Items[m.Cursor]
}
