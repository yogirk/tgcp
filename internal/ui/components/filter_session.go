package components

import tea "github.com/charmbracelet/bubbletea"

// FilterSession ties a filter model to a concrete list/table implementation.
// It centralizes applying a query, updating the table, and tracking counts.
type FilterSession[T any] struct {
	filter      *FilterModel
	allItems    []T
	getFiltered func([]T, string) []T
	updateTable func([]T)
}

// NewFilterSession creates a session that reuses the provided filter model.
func NewFilterSession[T any](
	filter *FilterModel,
	getFiltered func([]T, string) []T,
	updateTable func([]T),
) FilterSession[T] {
	return FilterSession[T]{
		filter:      filter,
		getFiltered: getFiltered,
		updateTable: updateTable,
	}
}

// Apply stores the full list, applies the current query, and updates counts.
func (s *FilterSession[T]) Apply(items []T) {
	s.allItems = items
	if s.filter == nil {
		s.updateTable(items)
		return
	}
	filtered := s.getFiltered(items, s.filter.Value())
	s.updateTable(filtered)
	s.filter.SetMatchCounts(len(items), len(filtered))
}

// HandleKey processes filter input and updates the table/counts as needed.
func (s *FilterSession[T]) HandleKey(msg tea.KeyMsg) FilterUpdateResult {
	if s.filter == nil {
		return FilterUpdateResult{Handled: false, ShouldContinue: true}
	}
	return HandleFilterUpdate(
		s.filter,
		msg,
		s.allItems,
		s.getFiltered,
		s.updateTable,
	)
}
