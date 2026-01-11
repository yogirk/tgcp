package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
)

// FilterModel provides a reusable filter input component with consistent behavior
// across all services. It handles the full filtering lifecycle including entering/exiting
// filter mode and managing the filter state.
type FilterModel struct {
	TextInput textinput.Model
	Active    bool
	Matches   int
	Total     int
}

// NewFilter creates a new FilterModel with default settings
func NewFilter() FilterModel {
	return NewFilterWithPlaceholder("Filter...")
}

// NewFilterWithPlaceholder creates a new FilterModel with a custom placeholder
func NewFilterWithPlaceholder(placeholder string) FilterModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 100
	ti.Width = 50
	ti.Prompt = "/ "

	return FilterModel{
		TextInput: ti,
		Active:    false,
	}
}

func (m FilterModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages for the filter input. When Active is true, it processes
// input. Returns a command to handle focus/blur and a boolean indicating if filter
// mode should remain active.
func (m FilterModel) Update(msg tea.Msg) (FilterModel, tea.Cmd) {
	if !m.Active {
		return m, nil
	}

	var cmd tea.Cmd
	m.TextInput, cmd = m.TextInput.Update(msg)
	return m, cmd
}

// View renders the filter input with clear state transitions:
// - Inactive, no filter: Subtle hint "Press / to filter"
// - Active: Prominent input with cursor
// - Inactive with filter: Badge showing applied filter
func (m FilterModel) View() string {
	query := m.TextInput.Value()

	// Style definitions
	labelStyle := styles.SubtleStyle
	hintStyle := styles.SubtleStyle
	countStyle := styles.SubtleStyle
	badgeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("232")).
		Background(styles.ColorBrandAccent).
		Padding(0, 1)
	activeInputStyle := lipgloss.NewStyle().
		Foreground(styles.ColorTextPrimary)
	sepStyle := lipgloss.NewStyle().
		Foreground(styles.ColorBorderSubtle)

	// Count text
	var countText string
	if query == "" {
		countText = fmt.Sprintf("Items: %d", m.Total)
	} else {
		countText = fmt.Sprintf("Matches: %d/%d", m.Matches, m.Total)
	}
	countView := countStyle.Render(countText)

	var bar string

	if m.Active {
		// ACTIVE STATE: Prominent input with cursor
		label := labelStyle.Render("Filter:")
		inputView := activeInputStyle.Render(m.TextInput.View())
		bar = lipgloss.JoinHorizontal(
			lipgloss.Left,
			label, " ", inputView, "  ", countView,
		)
	} else if query != "" {
		// INACTIVE WITH FILTER: Show badge + clear hint
		label := labelStyle.Render("Filter:")
		badge := badgeStyle.Render(query)
		sep := sepStyle.Render(" │ ")
		clearHint := hintStyle.Render("Esc to clear")
		bar = lipgloss.JoinHorizontal(
			lipgloss.Left,
			label, " ", badge, "  ", countView, sep, clearHint,
		)
	} else {
		// INACTIVE, NO FILTER: Subtle hint
		label := labelStyle.Render("Filter:")
		hint := hintStyle.Render("Press / to filter")
		bar = lipgloss.JoinHorizontal(
			lipgloss.Left,
			label, " ", hint, "  ", countView,
		)
	}

	return bar
}

// EnterFilterMode activates the filter and focuses the input
func (m *FilterModel) EnterFilterMode() tea.Cmd {
	m.Active = true
	return m.TextInput.Focus()
}

// ExitFilterMode deactivates the filter and resets the input
func (m *FilterModel) ExitFilterMode() {
	m.Active = false
	m.TextInput.Blur()
	m.TextInput.Reset()
}

// ExitFilterModeKeepValue deactivates the filter but keeps the current value
func (m *FilterModel) ExitFilterModeKeepValue() {
	m.Active = false
	m.TextInput.Blur()
}

// HandleKeyMsg processes key messages for filter mode. Returns:
// - shouldExit: true if filter mode should be exited
// - shouldKeepValue: true if filter value should be kept when exiting
// - cmd: command to execute (usually textinput.Blink when entering)
func (m *FilterModel) HandleKeyMsg(msg tea.KeyMsg) (shouldExit bool, shouldKeepValue bool, cmd tea.Cmd) {
	if !m.Active {
		// Check if we should enter filter mode
		if msg.String() == "/" {
			cmd = m.EnterFilterMode()
			return false, false, cmd
		}
		if msg.String() == "esc" && m.TextInput.Value() != "" {
			m.TextInput.Reset()
			return true, false, nil
		}
		return false, false, nil
	}

	// Handle exit keys
	switch msg.String() {
	case "esc":
		m.ExitFilterMode()
		return true, false, nil
	case "enter":
		m.ExitFilterModeKeepValue()
		return true, true, nil
	}

	// Update the input for other keys
	var inputCmd tea.Cmd
	m.TextInput, inputCmd = m.TextInput.Update(msg)
	return false, false, inputCmd
}

// IsNavigationKey checks if a key is a navigation key that should pass through to the table
func IsNavigationKey(key string) bool {
	return key == "up" || key == "down" || key == "j" || key == "k" || key == "g" || key == "G" ||
		key == "home" || key == "end" || key == "pageup" || key == "pagedown"
}

// Value returns the current filter query string
func (m FilterModel) Value() string {
	return m.TextInput.Value()
}

// IsActive returns whether the filter is currently active
func (m FilterModel) IsActive() bool {
	return m.Active
}

// SetMatchCounts updates the total and matched counts for the filter bar.
func (m *FilterModel) SetMatchCounts(total, matches int) {
	m.Total = total
	m.Matches = matches
}

// FilterSlice is a generic helper function that filters a slice of items based on
// a query string. The matcher function should return true if an item matches the query.
func FilterSlice[T any](items []T, query string, matcher func(item T, query string) bool) []T {
	if query == "" {
		return items
	}

	queryLower := strings.ToLower(query)
	var matches []T
	for _, item := range items {
		if matcher(item, queryLower) {
			matches = append(matches, item)
		}
	}
	return matches
}

// ContainsMatch is a helper matcher function that checks if any of the provided
// string fields contain the query (case-insensitive).
func ContainsMatch(fields ...string) func(string) bool {
	return func(query string) bool {
		for _, field := range fields {
			if strings.Contains(strings.ToLower(field), query) {
				return true
			}
		}
		return false
	}
}

// FilterUpdateResult represents the result of processing a filter key message
type FilterUpdateResult struct {
	Handled        bool // Whether the filter handled this message
	ShouldContinue bool // Whether the service should continue processing this key
	Cmd            tea.Cmd
}

// HandleFilterUpdate processes a key message for filter mode. This centralizes
// the common filter handling logic used across all services.
//
// Parameters:
//   - filter: The filter model to handle
//   - msg: The key message to process
//   - allItems: All items in the list (unfiltered)
//   - getFiltered: Function that returns filtered items based on query
//   - updateTable: Function that updates the table with the given items
//
// Returns a FilterUpdateResult indicating whether the message was handled and
// any command that should be executed.
func HandleFilterUpdate[T any](
	filter *FilterModel,
	msg tea.KeyMsg,
	allItems []T,
	getFiltered func([]T, string) []T,
	updateTable func([]T),
) FilterUpdateResult {
	key := msg.String()

	// If filter is active and this is a navigation key, let it pass through to table
	if filter.IsActive() && IsNavigationKey(key) {
		// Apply current filter to table (in case filter value changed)
		filteredItems := getFiltered(allItems, filter.Value())
		updateTable(filteredItems)
		filter.SetMatchCounts(len(allItems), len(filteredItems))
		return FilterUpdateResult{
			Handled:        false,
			ShouldContinue: true,
			Cmd:            nil,
		}
	}

	// Check filter state before calling HandleKeyMsg (since it modifies the state)
	wasActive := filter.IsActive()

	// Handle filter mode entry/exit
	shouldExit, shouldKeepValue, filterCmd := filter.HandleKeyMsg(msg)

	// If we got a command and filter wasn't active before, we're entering filter mode
	if filterCmd != nil && !wasActive {
		// Entering filter mode - return the command to focus
		return FilterUpdateResult{
			Handled:        true,
			ShouldContinue: false,
			Cmd:            filterCmd,
		}
	}

	if shouldExit {
		if !shouldKeepValue {
			// Reset table when clearing filter
			updateTable(allItems)
			filter.SetMatchCounts(len(allItems), len(allItems))
		}
		if shouldKeepValue {
			filteredItems := getFiltered(allItems, filter.Value())
			updateTable(filteredItems)
			filter.SetMatchCounts(len(allItems), len(filteredItems))
		}
		// Consume the exit key to avoid triggering other actions.
		return FilterUpdateResult{
			Handled:        true,
			ShouldContinue: false,
			Cmd:            nil,
		}
	}

	if filter.IsActive() {
		// Filter is active - HandleKeyMsg already updated the TextInput internally
		// Now apply the filter to the table immediately
		filteredItems := getFiltered(allItems, filter.Value())
		updateTable(filteredItems)
		filter.SetMatchCounts(len(allItems), len(filteredItems))
		// Return the command from HandleKeyMsg (inputCmd for blinking cursor)
		return FilterUpdateResult{
			Handled:        true,
			ShouldContinue: false,
			Cmd:            filterCmd,
		}
	}

	// Filter didn't handle this message
	return FilterUpdateResult{
		Handled:        false,
		ShouldContinue: true,
		Cmd:            nil,
	}
}
