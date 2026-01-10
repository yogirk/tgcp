package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yogirk/tgcp/internal/styles"
)

type ServiceItem struct {
	Name      string
	ShortName string
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

func NewSidebar() SidebarModel {
	return SidebarModel{
		Items: []ServiceItem{
			{Name: "Overview", ShortName: "overview", Active: true},
			{Name: "Compute Engine", ShortName: "gce"},
			{Name: "Disks", ShortName: "disks"},
			{Name: "Kubernetes Engine", ShortName: "gke"},
			{Name: "Cloud SQL", ShortName: "sql"},
			{Name: "IAM", ShortName: "iam"},
			{Name: "Cloud Run", ShortName: "run"},
			{Name: "Cloud Storage", ShortName: "gcs"},
			{Name: "BigQuery", ShortName: "bq"},
			{Name: "Networking", ShortName: "net"},
			{Name: "Pub/Sub", ShortName: "pubsub"},
			{Name: "Memorystore", ShortName: "redis"},
			{Name: "Spanner", ShortName: "spanner"},
			{Name: "Bigtable", ShortName: "bigtable"},
			{Name: "Dataflow", ShortName: "dataflow"},
			{Name: "Dataproc", ShortName: "dataproc"},
			{Name: "Firestore", ShortName: "firestore"},
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
		name := item.Name
		if item.IsComing {
			name += " *"
		}

		isSelected := m.Cursor == i

		var renderedItem string
		if isSelected {
			if m.Active {
				renderedItem = styles.SelectedItemStyle.Render(name)
			} else {
				// Selected but not focused (dimmed)
				renderedItem = styles.UnselectedItemStyle.Copy().Foreground(styles.ColorHighlight).Render(name)
			}
		} else {
			style := styles.UnselectedItemStyle
			if item.IsComing {
				style = style.Copy().Foreground(styles.ColorSubtext)
			}
			renderedItem = style.Render(name)
		}

		doc.WriteString(renderedItem)
		doc.WriteString("\n")
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
