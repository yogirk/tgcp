package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
)

// Category represents a group of services
type Category struct {
	Name     string
	Expanded bool
	Services []ServiceItem
}

type HomeMenuModel struct {
	TopItem    *ServiceItem // Top-level item (e.g., Overview) shown before categories
	Categories []Category
	Cursor     int  // Index in the flattened visible list
	IsFocused  bool

	// Screen dimensions for mouse click calculations (set by parent)
	ScreenWidth  int
	ScreenHeight int
}

func NewHomeMenu() HomeMenuModel {
	return HomeMenuModel{
		// Top-level item (not in any category)
		TopItem: &ServiceItem{Name: "Overview (Command Center)", ShortName: "overview", Active: true},
		Categories: []Category{
			{
				Name:     "Compute",
				Expanded: true,
				Services: []ServiceItem{
					{Name: "Compute Engine (GCE)", ShortName: "gce"},
					{Name: "Kubernetes Engine (GKE)", ShortName: "gke"},
					{Name: "Cloud Run", ShortName: "run"},
				},
			},
			{
				Name:     "Storage",
				Expanded: true,
				Services: []ServiceItem{
					{Name: "Cloud Storage (GCS)", ShortName: "gcs"},
					{Name: "Disks (Block Storage)", ShortName: "disks"},
				},
			},
			{
				Name:     "Databases",
				Expanded: true,
				Services: []ServiceItem{
					{Name: "Cloud SQL", ShortName: "sql"},
					{Name: "Spanner", ShortName: "spanner"},
					{Name: "Bigtable", ShortName: "bigtable"},
					{Name: "Memorystore (Redis)", ShortName: "redis"},
					{Name: "Firestore / Datastore", ShortName: "firestore"},
				},
			},
			{
				Name:     "Data & Analytics",
				Expanded: true,
				Services: []ServiceItem{
					{Name: "BigQuery", ShortName: "bq"},
					{Name: "Dataflow", ShortName: "dataflow"},
					{Name: "Dataproc", ShortName: "dataproc"},
					{Name: "Pub/Sub", ShortName: "pubsub"},
				},
			},
			{
				Name:     "Security & Networking",
				Expanded: true,
				Services: []ServiceItem{
					{Name: "IAM & Admin", ShortName: "iam"},
					{Name: "Secret Manager", ShortName: "secrets"},
					{Name: "VPC Network", ShortName: "net"},
				},
			},
			{
				Name:     "Observability",
				Expanded: true,
				Services: []ServiceItem{
					{Name: "Cloud Logging", ShortName: "logs"},
				},
			},
		},
		Cursor:    0,
		IsFocused: true,
	}
}

// menuItem represents an item in the flattened visible list
type menuItem struct {
	isTopItem     bool
	isCategory    bool
	categoryIndex int
	serviceIndex  int // -1 for category headers
}

// getVisibleItems returns the flattened list of visible items
func (m HomeMenuModel) getVisibleItems() []menuItem {
	var items []menuItem

	// Add top-level item first (e.g., Overview)
	if m.TopItem != nil {
		items = append(items, menuItem{
			isTopItem:     true,
			categoryIndex: -1,
			serviceIndex:  -1,
		})
	}

	for catIdx, cat := range m.Categories {
		// Add category header
		items = append(items, menuItem{
			isCategory:    true,
			categoryIndex: catIdx,
			serviceIndex:  -1,
		})
		// Add services if expanded
		if cat.Expanded {
			for svcIdx := range cat.Services {
				items = append(items, menuItem{
					isCategory:    false,
					categoryIndex: catIdx,
					serviceIndex:  svcIdx,
				})
			}
		}
	}
	return items
}

func (m HomeMenuModel) Init() tea.Cmd {
	return nil
}

func (m HomeMenuModel) Update(msg tea.Msg) (HomeMenuModel, tea.Cmd) {
	visibleItems := m.getVisibleItems()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(visibleItems)-1 {
				m.Cursor++
			}
		case " ": // Space to toggle category
			if m.Cursor < len(visibleItems) {
				item := visibleItems[m.Cursor]
				if item.isCategory {
					m.Categories[item.categoryIndex].Expanded = !m.Categories[item.categoryIndex].Expanded
				}
			}
		}

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			// Map click Y to visible item index
			// The view has: box padding (1) + "Services" title (1) + blank line (1) + items
			// Plus top-level item has a blank line after it
			if idx := m.getItemIndexFromY(msg.Y, visibleItems); idx >= 0 {
				m.Cursor = idx
				// If clicking on a category, toggle it
				if idx < len(visibleItems) && visibleItems[idx].isCategory {
					m.Categories[visibleItems[idx].categoryIndex].Expanded = !m.Categories[visibleItems[idx].categoryIndex].Expanded
				}
			}
		}
	}
	return m, nil
}

// getItemIndexFromY maps a screen Y coordinate to a visible item index
// The menu box is centered on screen, so we calculate its position first
func (m HomeMenuModel) getItemIndexFromY(screenY int, visibleItems []menuItem) int {
	if m.ScreenHeight == 0 {
		return -1 // Dimensions not set
	}

	// Calculate menu box dimensions
	// Box content: padding(1) + title(1) + blank(1) + items + padding(1)
	// Items: each item is 1 row, top item has extra blank after it
	itemCount := len(visibleItems)
	contentHeight := 3 + itemCount + 1 // header rows + items + top item's extra blank
	for _, item := range visibleItems {
		if item.isTopItem {
			contentHeight++ // extra blank after top item
			break
		}
	}

	// Menu is vertically centered (approximately - there's banner and info box above)
	// The landing page layout: banner + info + menu + hints, all centered
	// For a rough estimate, assume menu starts around 40% down the screen
	menuStartY := m.ScreenHeight * 2 / 5

	// Calculate relative Y within the menu
	relativeY := screenY - menuStartY

	// Account for box padding (1 row) and title + blank (2 rows)
	const headerOffset = 3 // padding + "Services" + blank line
	contentY := relativeY - headerOffset

	if contentY < 0 {
		return -1 // Click above content
	}

	// Map content Y to item index
	row := 0
	for i, item := range visibleItems {
		if contentY == row {
			return i
		}
		row++
		// Top item has extra blank line after it
		if item.isTopItem {
			row++
		}
	}

	return -1
}

func (m HomeMenuModel) View() string {
	visibleItems := m.getVisibleItems()
	var lines []string

	for i, item := range visibleItems {
		isSelected := m.Cursor == i

		if item.isTopItem {
			// Render top-level item (Overview)
			name := m.TopItem.Name
			var rendered string
			if isSelected {
				rendered = styles.SelectedItemStyle.Copy().
					UnsetBorderLeft().
					Render("▸ " + name)
			} else {
				rendered = styles.UnselectedItemStyle.Copy().
					PaddingLeft(2).
					Render(name)
			}
			lines = append(lines, rendered)
			lines = append(lines, "") // Add spacing after top item
		} else if item.isCategory {
			// Render category header
			cat := m.Categories[item.categoryIndex]
			arrow := "▼"
			if !cat.Expanded {
				arrow = "▶"
			}

			// Show count when collapsed
			label := cat.Name
			if !cat.Expanded {
				label = cat.Name + " (" + itoa(len(cat.Services)) + ")"
			}

			var rendered string
			if isSelected {
				rendered = styles.SelectedItemStyle.Copy().
					Bold(true).
					Render(arrow + " " + label)
			} else {
				rendered = lipgloss.NewStyle().
					Foreground(styles.ColorBrandAccent).
					Bold(true).
					PaddingLeft(1).
					Render(arrow + " " + label)
			}
			lines = append(lines, rendered)
		} else {
			// Render service item
			svc := m.Categories[item.categoryIndex].Services[item.serviceIndex]
			name := svc.Name
			if svc.IsComing {
				name += " [Coming Soon]"
			}

			var rendered string
			if isSelected {
				rendered = styles.SelectedItemStyle.Copy().
					UnsetBorderLeft().
					Render("  ▸ " + name)
			} else {
				style := styles.UnselectedItemStyle.Copy().PaddingLeft(4)
				if svc.IsComing {
					style = style.Foreground(styles.ColorTextMuted)
				}
				rendered = style.Render(name)
			}
			lines = append(lines, rendered)
		}
	}

	listContent := lipgloss.JoinVertical(lipgloss.Left, lines...)

	// Title
	title := styles.HeaderStyle.Render("Services")

	// Combine Title + Content
	content := lipgloss.JoinVertical(lipgloss.Left, title, "", listContent)

	// Wrap in a card/box
	menuBox := styles.BoxStyle.Copy().
		BorderForeground(styles.ColorTextMuted).
		Padding(1, 2).
		Render(content)

	return menuBox
}

// SelectedItem returns the currently selected service (or empty if on a category)
func (m HomeMenuModel) SelectedItem() ServiceItem {
	visibleItems := m.getVisibleItems()
	if m.Cursor >= len(visibleItems) {
		return ServiceItem{}
	}

	item := visibleItems[m.Cursor]
	if item.isTopItem {
		return *m.TopItem
	}
	if item.isCategory {
		return ServiceItem{} // No service selected when on category header
	}

	return m.Categories[item.categoryIndex].Services[item.serviceIndex]
}

// IsOnCategory returns true if cursor is on a category header
func (m HomeMenuModel) IsOnCategory() bool {
	visibleItems := m.getVisibleItems()
	if m.Cursor >= len(visibleItems) {
		return false
	}
	return visibleItems[m.Cursor].isCategory
}

// ToggleCurrentCategory toggles the category if cursor is on one
func (m *HomeMenuModel) ToggleCurrentCategory() {
	visibleItems := m.getVisibleItems()
	if m.Cursor >= len(visibleItems) {
		return
	}
	item := visibleItems[m.Cursor]
	if item.isCategory {
		m.Categories[item.categoryIndex].Expanded = !m.Categories[item.categoryIndex].Expanded
	}
}

// itoa is a simple int to string conversion
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
