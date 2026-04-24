package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"

	"github.com/yogirk/tgcp/internal/styles"
)

// Category represents a group of services
type Category struct {
	Name     string
	Expanded bool
	Services []ServiceItem
}

// serviceIcons maps service short names to icons (mirrors sidebar)
var serviceIcons = map[string]string{
	"overview":         "◉",
	"gce":              "⚙",
	"gke":              "☸",
	"run":              "▷",
	"gcs":              "▤",
	"disks":            "◔",
	"sql":              "⛁",
	"spanner":          "⬡",
	"bigtable":         "▦",
	"redis":            "◇",
	"firestore":        "◲",
	"bq":               "⊞",
	"dataflow":         "⇢",
	"dataproc":         "⎈",
	"pubsub":           "⇌",
	"iam":              "⚿",
	"secrets":          "✦",
	"net":              "⇄",
	"logs":             "☰",
	"cloudbuild":       "◈",
	"artifactregistry": "▣",
}

// listEntry represents a single row in the flat list (either category header or service)
type listEntry struct {
	isCategory   bool
	isTopItem    bool
	categoryIdx  int
	serviceIdx   int
	service      *ServiceItem // nil for category headers
	categoryName string
	searchText   string // "Name ShortName" for fuzzy matching
}

type HomeMenuModel struct {
	TopItem    *ServiceItem
	Categories []Category
	IsFocused  bool

	// Screen dimensions
	ScreenWidth  int
	ScreenHeight int

	filter       FilterModel
	allEntries   []listEntry // flat list: top item + categories + services
	filtered     []listEntry // after fuzzy filter
	cursor       int         // index into selectable items in filtered list
	scrollOffset int
	viewportRows int
	matchIndexes map[int][]int // allEntries index -> matched char positions
}

func NewHomeMenu() HomeMenuModel {
	m := HomeMenuModel{
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
			{
				Name:     "DevOps",
				Expanded: true,
				Services: []ServiceItem{
					{Name: "Cloud Build", ShortName: "cloudbuild"},
					{Name: "Artifact Registry", ShortName: "artifactregistry"},
				},
			},
		},
		IsFocused:    true,
		filter:       NewFilterWithPlaceholder("Type to filter services..."),
		viewportRows: 12, // conservative default before first WindowSizeMsg
	}
	m.rebuildEntries()
	m.applyFilter()
	return m
}

// rebuildEntries builds the flat list from TopItem + Categories
func (m *HomeMenuModel) rebuildEntries() {
	m.allEntries = nil

	if m.TopItem != nil {
		m.allEntries = append(m.allEntries, listEntry{
			isTopItem:  true,
			service:    m.TopItem,
			searchText: m.TopItem.Name + " " + m.TopItem.ShortName,
		})
	}

	for catIdx, cat := range m.Categories {
		m.allEntries = append(m.allEntries, listEntry{
			isCategory:   true,
			categoryIdx:  catIdx,
			categoryName: cat.Name,
		})
		for svcIdx := range cat.Services {
			svc := &m.Categories[catIdx].Services[svcIdx]
			m.allEntries = append(m.allEntries, listEntry{
				categoryIdx: catIdx,
				serviceIdx:  svcIdx,
				service:     svc,
				searchText:  svc.Name + " " + svc.ShortName,
			})
		}
	}
}

// applyFilter filters the flat list based on the current filter query
func (m *HomeMenuModel) applyFilter() {
	query := m.filter.Value()
	m.matchIndexes = make(map[int][]int)

	if query == "" {
		// No filter — show everything
		m.filtered = make([]listEntry, len(m.allEntries))
		copy(m.filtered, m.allEntries)
		m.updateFilterCounts()
		m.clampCursorAndScroll()
		return
	}

	// Collect searchable entries (services only) with their allEntries indices
	type searchable struct {
		idx  int
		text string
	}
	var targets []searchable
	var texts []string
	for i, e := range m.allEntries {
		if e.service != nil {
			targets = append(targets, searchable{idx: i, text: e.searchText})
			texts = append(texts, e.searchText)
		}
	}

	// Run fuzzy match
	matches := fuzzy.Find(query, texts)
	matchedEntryIdxs := make(map[int]bool)
	for _, match := range matches {
		entryIdx := targets[match.Index].idx
		matchedEntryIdxs[entryIdx] = true
		m.matchIndexes[entryIdx] = match.MatchedIndexes
	}

	// Build filtered list: include category headers only if they have matching services
	m.filtered = nil
	for i, e := range m.allEntries {
		if e.isCategory {
			if m.categoryHasMatch(e.categoryIdx, matchedEntryIdxs) {
				m.filtered = append(m.filtered, e)
			}
		} else if matchedEntryIdxs[i] {
			m.filtered = append(m.filtered, e)
		}
	}

	// Reset cursor to first match when filtering
	m.cursor = 0
	m.scrollOffset = 0
	m.updateFilterCounts()
	m.clampCursorAndScroll()
}

// categoryHasMatch checks if any service in the given category matched the filter
func (m *HomeMenuModel) categoryHasMatch(catIdx int, matchedIdxs map[int]bool) bool {
	for i, e := range m.allEntries {
		if !e.isCategory && e.categoryIdx == catIdx && matchedIdxs[i] {
			return true
		}
	}
	return false
}

// updateFilterCounts sets the filter bar match/total counts
func (m *HomeMenuModel) updateFilterCounts() {
	// Count total selectable items (services)
	total := 0
	for _, e := range m.allEntries {
		if e.service != nil {
			total++
		}
	}
	matched := 0
	for _, e := range m.filtered {
		if e.service != nil {
			matched++
		}
	}
	m.filter.SetMatchCounts(total, matched)
}

// selectableItems returns only selectable entries from the filtered list
func (m *HomeMenuModel) selectableItems() []int {
	var indices []int
	for i, e := range m.filtered {
		if e.service != nil {
			indices = append(indices, i)
		}
	}
	return indices
}

// clampCursorAndScroll ensures cursor and scroll are within valid bounds
func (m *HomeMenuModel) clampCursorAndScroll() {
	selectable := m.selectableItems()
	if len(selectable) == 0 {
		m.cursor = 0
		m.scrollOffset = 0
		return
	}
	if m.cursor >= len(selectable) {
		m.cursor = len(selectable) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.adjustScroll()
}

// adjustScroll ensures the cursor row is visible in the viewport
func (m *HomeMenuModel) adjustScroll() {
	selectable := m.selectableItems()
	if len(selectable) == 0 || m.viewportRows <= 0 {
		return
	}

	// The cursor's actual row in the filtered list
	cursorRow := 0
	if m.cursor < len(selectable) {
		cursorRow = selectable[m.cursor]
	}

	// Ensure cursor row is visible
	if cursorRow < m.scrollOffset {
		m.scrollOffset = cursorRow
	}
	if cursorRow >= m.scrollOffset+m.viewportRows {
		m.scrollOffset = cursorRow - m.viewportRows + 1
	}

	// Clamp scroll offset
	maxScroll := len(m.filtered) - m.viewportRows
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.scrollOffset > maxScroll {
		m.scrollOffset = maxScroll
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

// isArrowNavKey returns true for non-printable navigation keys that should
// pass through to cursor movement even when the filter input has focus.
func isArrowNavKey(key string) bool {
	return key == "up" || key == "down" ||
		key == "home" || key == "end" ||
		key == "pageup" || key == "pagedown"
}

func (m HomeMenuModel) Init() tea.Cmd {
	return nil
}

func (m HomeMenuModel) Update(msg tea.Msg) (HomeMenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		// Clear applied filter with q or esc (when filter is not active but has a value)
		if !m.filter.Active && m.filter.Value() != "" && (key == "q" || key == "esc") {
			m.filter.TextInput.Reset()
			m.applyFilter()
			return m, nil
		}

		// Let filter handle its lifecycle (enter/exit/typing)
		if m.filter.Active || key == "/" {
			shouldExit, _, cmd := m.filter.HandleKeyMsg(msg)

			if shouldExit {
				m.applyFilter()
				return m, cmd
			}

			if m.filter.Active {
				// Only non-printable nav keys fall through; printable keys stay in filter
				if isArrowNavKey(key) {
					m.applyFilter()
					// fall through to navigation below
				} else {
					m.applyFilter()
					return m, cmd
				}
			} else if cmd != nil {
				// Just entered filter mode
				return m, cmd
			}
		}

		// Navigation (works whether filter is active or not)
		selectable := m.selectableItems()
		switch key {
		case "up":
			if m.cursor > 0 {
				m.cursor--
				m.adjustScroll()
			}
		case "k":
			if !m.filter.Active && m.cursor > 0 {
				m.cursor--
				m.adjustScroll()
			}
		case "down":
			if m.cursor < len(selectable)-1 {
				m.cursor++
				m.adjustScroll()
			}
		case "j":
			if !m.filter.Active && m.cursor < len(selectable)-1 {
				m.cursor++
				m.adjustScroll()
			}
		case "home", "g":
			if !m.filter.Active {
				m.cursor = 0
				m.adjustScroll()
			}
		case "end", "G":
			if !m.filter.Active && len(selectable) > 0 {
				m.cursor = len(selectable) - 1
				m.adjustScroll()
			}
		}

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			// Map click Y to a filtered list entry
			if idx := m.getItemFromClickY(msg.Y); idx >= 0 {
				// Find which selectable index this corresponds to
				selectable := m.selectableItems()
				for si, fi := range selectable {
					if fi == idx {
						m.cursor = si
						break
					}
				}
			}
		}
	}
	return m, nil
}

// getItemFromClickY maps a screen Y to a filtered list index (best effort)
func (m HomeMenuModel) getItemFromClickY(screenY int) int {
	// The menu is rendered inside a centered layout. We estimate:
	// filter bar takes 1 row, then blank row, then list items
	// This is approximate since we don't know the exact placement.
	// We'll treat relative Y = screenY offset from estimated content start.
	// For now, use scroll offset to map.
	const filterBarRows = 2 // filter bar + blank line
	const boxPadding = 1    // top padding of box
	const titleRows = 2     // "Services" + blank line

	// Rough estimate of where the menu box content starts
	// The landing page centers content vertically, so estimate ~40% down
	menuStartY := m.ScreenHeight * 2 / 5
	relY := screenY - menuStartY - boxPadding - titleRows - filterBarRows

	if relY < 0 {
		return -1
	}

	targetRow := m.scrollOffset + relY
	if targetRow >= 0 && targetRow < len(m.filtered) {
		return targetRow
	}
	return -1
}

func (m HomeMenuModel) View() string {
	// Filter bar
	filterBar := m.filter.View()

	// Title
	title := styles.HeaderStyle.Render("Services")

	// Category header style (muted uppercase divider)
	catStyle := styles.GroupStyle

	// Visible slice
	endIdx := m.scrollOffset + m.viewportRows
	if endIdx > len(m.filtered) {
		endIdx = len(m.filtered)
	}
	startIdx := m.scrollOffset
	if startIdx < 0 {
		startIdx = 0
	}

	// Build the selectable items map for cursor highlighting
	selectable := m.selectableItems()
	cursorFilteredIdx := -1
	if m.cursor >= 0 && m.cursor < len(selectable) {
		cursorFilteredIdx = selectable[m.cursor]
	}

	var lines []string
	for i := startIdx; i < endIdx; i++ {
		entry := m.filtered[i]
		if entry.isCategory {
			// Category header: uppercase, muted, non-selectable
			header := catStyle.Render(strings.ToUpper(entry.categoryName))
			lines = append(lines, header)
		} else if entry.service != nil {
			// Service item
			icon := serviceIcons[entry.service.ShortName]
			if icon == "" {
				icon = "·"
			}
			name := entry.service.Name
			if entry.service.IsComing {
				name += " [Coming Soon]"
			}

			display := icon + "  " + name
			isSelected := i == cursorFilteredIdx

			if isSelected {
				rendered := styles.SelectedActive.Render(display)
				lines = append(lines, rendered)
			} else {
				style := styles.UnselectedItemStyle
				if entry.service.IsComing {
					style = style.Copy().Foreground(styles.ColorTextMuted)
				}
				lines = append(lines, style.Render(display))
			}
		}
	}

	// Scroll indicators
	var scrollUp, scrollDown string
	if m.scrollOffset > 0 {
		scrollUp = styles.SubtleStyle.Render("  ↑ more")
	}
	if m.scrollOffset+m.viewportRows < len(m.filtered) {
		scrollDown = styles.SubtleStyle.Render("  ↓ more")
	}

	// Combine list with scroll indicators
	listContent := lipgloss.JoinVertical(lipgloss.Left, lines...)
	if scrollUp != "" {
		listContent = lipgloss.JoinVertical(lipgloss.Left, scrollUp, listContent)
	}
	if scrollDown != "" {
		listContent = lipgloss.JoinVertical(lipgloss.Left, listContent, scrollDown)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, title, "", filterBar, "", listContent)

	// Fixed-height box so the menu doesn't grow beyond the viewport
	boxHeight := m.viewportRows + 8 // items + title(1) + gaps(2) + filter(1) + padding(2) + borders(2)
	menuBox := styles.PrimaryBoxStyle.Copy().
		Height(boxHeight).
		Render(content)

	return menuBox
}

// SelectedItem returns the currently selected service
func (m HomeMenuModel) SelectedItem() ServiceItem {
	selectable := m.selectableItems()
	if m.cursor < 0 || m.cursor >= len(selectable) {
		return ServiceItem{}
	}
	entry := m.filtered[selectable[m.cursor]]
	if entry.service != nil {
		return *entry.service
	}
	return ServiceItem{}
}

// IsOnCategory returns true if cursor is on a category header.
// With the new list, cursor never lands on headers.
func (m HomeMenuModel) IsOnCategory() bool {
	return false
}

// ToggleCurrentCategory is a no-op with the new flat list.
func (m *HomeMenuModel) ToggleCurrentCategory() {}

// FilterActive returns whether the filter input is currently active
func (m HomeMenuModel) FilterActive() bool {
	return m.filter.IsActive()
}

// HasFilter returns whether a non-empty filter value is applied
func (m HomeMenuModel) HasFilter() bool {
	return m.filter.Value() != ""
}

// UpdateViewportRows recalculates the number of visible rows from screen height.
// The menu box must stay compact enough that the banner, info box, hints, and
// status bar all remain visible on screen.
func (m *HomeMenuModel) UpdateViewportRows() {
	// External overhead (outside the menu box):
	//   banner: 6, gap: 1, info box: 3, gap: 1, gap: 1, hints: 1, version: 1, status bar: 1 = 15
	// Internal box overhead (borders, padding, title, filter):
	//   border: 2, padding: 2, title+gap: 2, filter+gap: 2, scroll indicator: 1 = 9
	const overhead = 24
	rows := m.ScreenHeight - overhead
	if rows < 5 {
		rows = 5
	}
	// Cap at 15 rows so the list stays compact and scrollable
	if rows > 15 {
		rows = 15
	}
	m.viewportRows = rows
	m.clampCursorAndScroll()
}
