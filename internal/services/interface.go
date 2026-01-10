package services

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

// Service represents a pluggable GCP service module
type Service interface {
	// Name returns the display name (e.g. "Google Compute Engine")
	Name() string

	// ShortName returns the ID for command palette (e.g. "gce")
	ShortName() string

	// InitService initializes the service (API clients, empty state)
	InitService(ctx context.Context, projectID string) error

	// Reinit reinitializes the service with a new project ID
	// This is called when switching projects and should reset state and reinitialize clients
	Reinit(ctx context.Context, projectID string) error

	// Update handles messages specific to this service
	Update(msg tea.Msg) (tea.Model, tea.Cmd)

	// View renders the service UI
	View() string

	// HelpText returns the context-aware help text for the status bar
	HelpText() string

	// Refresh triggers a data reload
	Refresh() tea.Cmd

	// Focus is called when the service gains input focus
	Focus()

	// Blur is called when the service loses input focus
	Blur()

	// Reset resets the service state (e.g. back to list view)
	Reset()

	// IsRootView returns true if the service is at its top-level view (e.g. List)
	// Used to determine if 'q' should exit the service or go back
	IsRootView() bool
}
