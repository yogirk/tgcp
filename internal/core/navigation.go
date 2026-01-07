package core

import (
	"github.com/sahilm/fuzzy"
)

// ViewType enumerates available views
type ViewType int

const (
	ViewHome ViewType = iota
	ViewServiceList
	ViewResourceDetail
	ViewHelp
)

// Route represents a navigational destination
type Route struct {
	View    ViewType
	Service string // e.g., "gce", "sql"
	ID      string // resource ID
}

// Command represents an actionable command in the palette
type Command struct {
	Name        string
	Description string
	Action      func() Route
}

// NavigationModel manages routing and command palette state
type NavigationModel struct {
	CurrentRoute Route
	History      []Route
	Commands     []Command

	// Palette State
	PaletteActive bool
	Query         string
	Suggestions   []Command
	Selection     int
}

func NewNavigation() NavigationModel {
	return NavigationModel{
		CurrentRoute:  Route{View: ViewHome},
		History:       make([]Route, 0),
		Commands:      defaultCommands(),
		PaletteActive: false,
		Suggestions:   []Command{},
	}
}

func defaultCommands() []Command {
	return []Command{
		{Name: "Home", Description: "Go to Home Screen", Action: func() Route { return Route{View: ViewHome} }},
		{Name: "GCE: List Instances", Description: "List Google Compute Engine VM instances", Action: func() Route { return Route{View: ViewServiceList, Service: "gce"} }},
		{Name: "SQL: List Instances", Description: "List Cloud SQL instances", Action: func() Route { return Route{View: ViewServiceList, Service: "sql"} }},
		{Name: "IAM: List Service Accounts", Description: "List IAM Service Accounts", Action: func() Route { return Route{View: ViewServiceList, Service: "iam"} }},
		{Name: "Help", Description: "Show Help Screen", Action: func() Route { return Route{View: ViewHelp} }},
		// Add more commands here (e.g., Refresh, Quit)
	}
}

// FilterCommands updates suggestions based on input query
func (m *NavigationModel) FilterCommands(query string) {
	m.Query = query
	if query == "" {
		m.Suggestions = []Command{}
		return
	}

	// Prepare source for fuzzy search
	sources := make([]string, len(m.Commands))
	for i, cmd := range m.Commands {
		sources[i] = cmd.Name + " " + cmd.Description
	}

	matches := fuzzy.Find(query, sources)

	m.Suggestions = make([]Command, 0, len(matches))
	for _, match := range matches {
		m.Suggestions = append(m.Suggestions, m.Commands[match.Index])
	}
	m.Selection = 0 // Reset selection
}

// SelectNext moves selection down
func (m *NavigationModel) SelectNext() {
	if m.Selection < len(m.Suggestions)-1 {
		m.Selection++
	}
}

// SelectPrev moves selection up
func (m *NavigationModel) SelectPrev() {
	if m.Selection > 0 {
		m.Selection--
	}
}

// ExecuteSelection returns the route for the selected command
func (m *NavigationModel) ExecuteSelection() *Route {
	if len(m.Suggestions) > 0 {
		route := m.Suggestions[m.Selection].Action()
		return &route
	}
	return nil
}
