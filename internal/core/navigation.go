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
	ViewProjectSwitcher
)

// Route represents a navigational destination
type Route struct {
	View    ViewType
	Service string // e.g., "gce", "sql"
	ID      string // resource ID or Project ID
}

// Command represents an actionable command in the palette
type Command struct {
	Name        string
	Description string
	Action      func() Route
}

// SuggestionMatch wraps a Command with fuzzy match info for highlighting
type SuggestionMatch struct {
	Command
	MatchedIndexes []int // Positions of matched characters in Name+Description
}

// NavigationModel manages routing and command palette state
type NavigationModel struct {
	CurrentRoute Route
	History      []Route
	Commands     []Command
	BaseCommands []Command // Persist default commands

	// Palette State
	PaletteActive bool
	Query         string
	Suggestions   []SuggestionMatch // Includes match info for highlighting
	Selection     int
}

func NewNavigation() NavigationModel {
	defaults := defaultCommands()
	return NavigationModel{
		CurrentRoute:  Route{View: ViewHome},
		History:       make([]Route, 0),
		Commands:      defaults,
		BaseCommands:  defaults,
		PaletteActive: false,
		Suggestions:   []SuggestionMatch{},
	}
}

func defaultCommands() []Command {
	return []Command{
		{Name: "GCP: Switch Project", Description: "Switch active Google Cloud Project", Action: func() Route { return Route{View: ViewProjectSwitcher} }},
		{Name: "Home", Description: "Go to Home Screen", Action: func() Route { return Route{View: ViewHome} }},

		// Compute
		{Name: "GCE: List Instances", Description: "List Google Compute Engine VM instances", Action: func() Route { return Route{View: ViewServiceList, Service: "gce"} }},
		{Name: "GKE: List Clusters", Description: "List Kubernetes Engine Clusters", Action: func() Route { return Route{View: ViewServiceList, Service: "gke"} }},
		{Name: "Disks: List Disks", Description: "List Persistent Disks (Block Storage)", Action: func() Route { return Route{View: ViewServiceList, Service: "disks"} }},
		{Name: "Run: List Services", Description: "List Cloud Run Services", Action: func() Route { return Route{View: ViewServiceList, Service: "run"} }},

		// Data & Storage
		{Name: "SQL: List Instances", Description: "List Cloud SQL instances", Action: func() Route { return Route{View: ViewServiceList, Service: "sql"} }},
		{Name: "GCS: List Buckets", Description: "List Cloud Storage Buckets", Action: func() Route { return Route{View: ViewServiceList, Service: "gcs"} }},
		{Name: "BigQuery: List Datasets", Description: "List BigQuery Datasets", Action: func() Route { return Route{View: ViewServiceList, Service: "bq"} }},
		{Name: "Redis: List Instances", Description: "List Memorystore (Redis) Instances", Action: func() Route { return Route{View: ViewServiceList, Service: "redis"} }},
		{Name: "Spanner: List Instances", Description: "List Spanner Instances", Action: func() Route { return Route{View: ViewServiceList, Service: "spanner"} }},
		{Name: "Bigtable: List Instances", Description: "List Bigtable Instances", Action: func() Route { return Route{View: ViewServiceList, Service: "bigtable"} }},
		{Name: "Firestore: List Databases", Description: "List Firestore Databases", Action: func() Route { return Route{View: ViewServiceList, Service: "firestore"} }},

		// Messaging & Processing
		{Name: "Pub/Sub: List Topics", Description: "List Pub/Sub Topics", Action: func() Route { return Route{View: ViewServiceList, Service: "pubsub"} }},
		{Name: "Dataflow: List Jobs", Description: "List Dataflow Jobs", Action: func() Route { return Route{View: ViewServiceList, Service: "dataflow"} }},
		{Name: "Dataproc: List Clusters", Description: "List Dataproc Clusters", Action: func() Route { return Route{View: ViewServiceList, Service: "dataproc"} }},

		// Management
		{Name: "IAM: List Service Accounts", Description: "List IAM Service Accounts", Action: func() Route { return Route{View: ViewServiceList, Service: "iam"} }},
		{Name: "VPC: List Networks", Description: "List VPC Networks", Action: func() Route { return Route{View: ViewServiceList, Service: "net"} }},

		{Name: "Help", Description: "Show Help Screen", Action: func() Route { return Route{View: ViewHelp} }},
	}
}

// SetCommands updates the available commands (e.g. for switching context)
func (m *NavigationModel) SetCommands(cmds []Command) {
	m.Commands = cmds
	m.Selection = 0
	m.FilterCommands("") // Reset filter
}

// RestoreBaseCommands resets to default commands
func (m *NavigationModel) RestoreBaseCommands() {
	m.Commands = m.BaseCommands
	m.Selection = 0
	m.FilterCommands("")
}

// FilterCommands updates suggestions based on input query
func (m *NavigationModel) FilterCommands(query string) {
	m.Query = query
	if query == "" {
		m.Suggestions = []SuggestionMatch{}
		return
	}

	// Prepare source for fuzzy search
	sources := make([]string, len(m.Commands))
	for i, cmd := range m.Commands {
		sources[i] = cmd.Name + " " + cmd.Description
	}

	matches := fuzzy.Find(query, sources)

	m.Suggestions = make([]SuggestionMatch, 0, len(matches))
	for _, match := range matches {
		m.Suggestions = append(m.Suggestions, SuggestionMatch{
			Command:        m.Commands[match.Index],
			MatchedIndexes: match.MatchedIndexes,
		})
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
