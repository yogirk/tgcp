package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
)

// Standard colors for table selection
const (
	TableSelectedFocused = lipgloss.Color("236") // Dark grey background when focused
	TableSelectedBlurred = lipgloss.Color("240") // Lighter grey background when blurred
	TableTextFocused     = lipgloss.Color("39")  // Brand accent (light blue) when focused
	TableTextBlurred     = lipgloss.Color("245") // Muted text when blurred
	TableHeaderBg        = lipgloss.Color("237") // Subtle background for headers
)

// StandardTable is a standardized table component with built-in Focus/Blur and window size handling
type StandardTable struct {
	table.Model
	focused     bool
	heightOffset int
}

// TableOption is a function that configures a StandardTable
type TableOption func(*StandardTable)

// WithHeight sets the initial height of the table
func WithHeight(height int) TableOption {
	return func(t *StandardTable) {
		t.Model.SetHeight(height)
	}
}

// WithHeightOffset sets the height offset for window size calculations
func WithHeightOffset(offset int) TableOption {
	return func(t *StandardTable) {
		t.heightOffset = offset
	}
}

// WithFocused sets whether the table starts focused
func WithFocused(focused bool) TableOption {
	return func(t *StandardTable) {
		t.focused = focused
		if focused {
			t.Model.Focus()
		} else {
			t.Model.Blur()
		}
		t.applyStyles()
	}
}

// NewStandardTable creates a new standardized table with TGCP styling
func NewStandardTable(columns []table.Column, opts ...TableOption) *StandardTable {
	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	st := &StandardTable{
		Model:       t,
		focused:     true,
		heightOffset: 6, // Default offset
	}

	// Apply options
	for _, opt := range opts {
		opt(st)
	}

	// Apply initial styles
	st.applyStyles()

	return st
}

// applyStyles applies the appropriate styles based on focus state
func (st *StandardTable) applyStyles() {
	s := table.DefaultStyles()

	// Header style: subtle background, bold, primary text
	s.Header = lipgloss.NewStyle().
		Foreground(styles.ColorTextPrimary).
		Background(TableHeaderBg).
		Bold(true).
		Padding(0, 1)

	if st.focused {
		// Focused: Dark grey background, accent text, bold
		s.Selected = lipgloss.NewStyle().
			Foreground(TableTextFocused).
			Background(TableSelectedFocused).
			Bold(true)
	} else {
		// Blurred: Lighter grey background, muted text
		s.Selected = lipgloss.NewStyle().
			Foreground(TableTextBlurred).
			Background(TableSelectedBlurred).
			Bold(false)
	}

	st.Model.SetStyles(s)
}

// Focus sets focus and applies focused styling
func (st *StandardTable) Focus() {
	st.Model.Focus()
	st.focused = true
	st.applyStyles()
}

// Blur removes focus and applies blurred styling
func (st *StandardTable) Blur() {
	st.Model.Blur()
	st.focused = false
	st.applyStyles()
}

// HandleWindowSize calculates and sets appropriate height based on window size
func (st *StandardTable) HandleWindowSize(msg tea.WindowSizeMsg, heightOffset int) {
	newHeight := msg.Height - heightOffset
	if newHeight < 5 {
		newHeight = 5 // Minimum height
	}
	st.SetHeight(newHeight)
}

// HandleWindowSizeDefault uses the table's default height offset
func (st *StandardTable) HandleWindowSizeDefault(msg tea.WindowSizeMsg) {
	st.HandleWindowSize(msg, st.heightOffset)
}

// SetRows sets the table rows
func (st *StandardTable) SetRows(rows []table.Row) {
	st.Model.SetRows(rows)
	if len(rows) == 0 {
		return
	}
	cursor := st.Model.Cursor()
	if cursor < 0 || cursor >= len(rows) {
		st.Model.SetCursor(0)
	}
}

// SetHeight sets the table height
func (st *StandardTable) SetHeight(height int) {
	st.Model.SetHeight(height)
}

// SetCursor sets the cursor position
func (st *StandardTable) SetCursor(index int) {
	st.Model.SetCursor(index)
}

// Update handles messages and returns commands
func (st *StandardTable) Update(msg tea.Msg) (*StandardTable, tea.Cmd) {
	var cmd tea.Cmd
	st.Model, cmd = st.Model.Update(msg)
	return st, cmd
}

// View renders the table
func (st *StandardTable) View() string {
	return st.Model.View()
}

// -----------------------------------------------------------------------------
// Legacy TableModel (kept for backward compatibility)
// -----------------------------------------------------------------------------

type TableModel struct {
	Table table.Model
}

func NewTable(columns []table.Column, rows []table.Row) TableModel {
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = lipgloss.NewStyle().
		Foreground(styles.ColorTextPrimary).
		Background(TableHeaderBg).
		Bold(true).
		Padding(0, 1)
	s.Selected = s.Selected.
		Foreground(TableTextFocused).
		Background(TableSelectedFocused).
		Bold(true)
	t.SetStyles(s)

	return TableModel{Table: t}
}

func (m *TableModel) Update(msg tea.Msg) (TableModel, tea.Cmd) {
	var cmd tea.Cmd
	m.Table, cmd = m.Table.Update(msg)
	return *m, cmd
}

func (m TableModel) View() string {
	baseStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	return baseStyle.Render(m.Table.View())
}

func (m *TableModel) SetRows(rows []table.Row) {
	m.Table.SetRows(rows)
}

func (m *TableModel) SetHeight(h int) {
	m.Table.SetHeight(h)
}
