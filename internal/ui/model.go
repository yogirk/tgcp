package ui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rk/tgcp/internal/core"
	"github.com/rk/tgcp/internal/services"
	"github.com/rk/tgcp/internal/services/bigquery"
	"github.com/rk/tgcp/internal/services/cloudrun"
	"github.com/rk/tgcp/internal/services/cloudsql"
	"github.com/rk/tgcp/internal/services/gce"
	"github.com/rk/tgcp/internal/services/gcs"
	"github.com/rk/tgcp/internal/services/iam"
	"github.com/rk/tgcp/internal/ui/components"
)

// ViewMode defines the high-level view state
type ViewMode int

const (
	ViewHome ViewMode = iota
	ViewService
)

// FocusArea defines where the user input is directed
type FocusArea int

const (
	FocusSidebar FocusArea = iota
	FocusMain
	FocusPalette
)

// MainModel is the main application state
type MainModel struct {
	AuthState  core.AuthState
	Navigation core.NavigationModel

	// Layout
	Width  int
	Height int

	// Services
	ServiceMap map[string]services.Service
	CurrentSvc services.Service

	// Components
	Sidebar   components.SidebarModel
	HomeMenu  components.HomeMenuModel // Added
	StatusBar components.StatusBarModel
	Palette   components.PaletteModel // Added

	// State
	ViewMode      ViewMode // Added
	Focus         FocusArea
	LastFocus     FocusArea
	ShowHelp      bool
	ActiveService string
}

// InitialModel returns the initial state of the application
func InitialModel(authState core.AuthState) MainModel {
	// Initialize Cache
	cache := core.NewCache()

	// Initialize Services
	svcMap := make(map[string]services.Service)

	// Create GCE Service
	gceSvc := gce.NewService(cache)
	if authState.ProjectID != "" {
		gceSvc.InitService(context.Background(), authState.ProjectID)
	}
	svcMap["gce"] = gceSvc

	// Create Cloud SQL Service
	sqlSvc := cloudsql.NewService(cache)
	if authState.ProjectID != "" {
		sqlSvc.InitService(context.Background(), authState.ProjectID)
	}
	svcMap["sql"] = sqlSvc

	// Create IAM Service
	iamSvc := iam.NewService(cache)
	if authState.ProjectID != "" {
		iamSvc.InitService(context.Background(), authState.ProjectID)
	}
	svcMap["iam"] = iamSvc

	// Create Cloud Run Service
	runSvc := cloudrun.NewService(cache)
	if authState.ProjectID != "" {
		runSvc.InitService(context.Background(), authState.ProjectID)
	}
	svcMap["run"] = runSvc

	// Create GCS Service
	gcsSvc := gcs.NewService(cache)
	if authState.ProjectID != "" {
		gcsSvc.InitService(context.Background(), authState.ProjectID)
	}
	svcMap["gcs"] = gcsSvc

	// Create BigQuery Service
	bqSvc := bigquery.NewService(cache)
	if authState.ProjectID != "" {
		bqSvc.InitService(context.Background(), authState.ProjectID)
	}
	svcMap["bq"] = bqSvc

	return MainModel{
		AuthState:  authState,
		Navigation: core.NewNavigation(),
		Sidebar:    components.NewSidebar(),
		HomeMenu:   components.NewHomeMenu(), // Added
		StatusBar:  components.NewStatusBar(),
		Palette:    components.NewPalette(), // Added
		Focus:      FocusSidebar,
		ViewMode:   ViewHome, // Start at Home
		ServiceMap: svcMap,
	}
}

// Init initializes the bubbletea program
func (m MainModel) Init() tea.Cmd {
	return tea.EnableMouseCellMotion
}

// Update handles messages and updates the model
func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	// Status Bar Updates
	case core.StatusMsg:
		m.StatusBar.Message = msg.Message
		m.StatusBar.IsError = msg.IsError
		return m, nil

	case core.LastUpdatedMsg:
		m.StatusBar.LastUpdated = time.Time(msg)
		return m, nil

	case tea.KeyMsg:
		// Global Keybindings
		if m.Focus != FocusPalette {
			switch msg.String() {
			case "q":
				if m.ShowHelp {
					m.ShowHelp = false
					return m, nil
				}
				// Removed old global handler, new one is inside Service Mode block
				if m.ViewMode == ViewHome {
					return m, tea.Quit
				}
				// If in Service Mode, we delegate to the service specific block below
				// or if IsRootView logic there handles it.
				// But wait, if we are in Service Mode and IsRootView is true, we want Home.
				// The block below handles it. But if we are here, we need to NOT Quit.
				// Effectively, if ViewService, don't do anything here, let next block handle.
				if m.ViewMode == ViewService {
					// Fallthrough to Service Loop
				} else {
					return m, tea.Quit
				}
			case "ctrl+c":
				return m, tea.Quit
			case ":":
				m.LastFocus = m.Focus
				m.Focus = FocusPalette
				m.Navigation.PaletteActive = true
				m.StatusBar.Mode = "COMMAND"
				m.StatusBar.Message = "Type command..."
				return m, nil
			case "?":
				m.ShowHelp = !m.ShowHelp
				return m, nil
			case "tab":
				if m.ViewMode == ViewService {
					m.Sidebar.Visible = !m.Sidebar.Visible
					// Adjust focus if hiding active sidebar
					if !m.Sidebar.Visible && m.Focus == FocusSidebar {
						m.Focus = FocusMain
						m.Sidebar.Active = false
					}
				}
				return m, nil
			}
		} else {
			// Palette specific keys (Esc to close)
			// Palette specific keys (Esc to close)
			switch msg.String() {
			case "esc":
				m.Focus = m.LastFocus
				m.Navigation.PaletteActive = false
				m.StatusBar.Mode = "NORMAL"
				m.StatusBar.Message = "Ready"
				m.Palette.TextInput.Reset() // Clear input
				m.Navigation.FilterCommands("")
				return m, nil
			case "up":
				m.Navigation.SelectPrev()
				return m, nil
			case "down":
				m.Navigation.SelectNext()
				return m, nil
			case "enter":
				// Execute Command
				if route := m.Navigation.ExecuteSelection(); route != nil {
					// Route Logic
					if route.View == core.ViewHome {
						m.ViewMode = ViewHome
						m.Sidebar.Active = false
					} else if route.View == core.ViewServiceList {
						// Logic to switch service
						m.ViewMode = ViewService
						m.ActiveService = route.Service
						// Sync Sidebar
						for i, item := range m.Sidebar.Items {
							if item.ShortName == route.Service {
								m.Sidebar.Cursor = i
							}
						}
						// Initialize service if needed (similar to sidebar logic)
						if svc, exists := m.ServiceMap[m.ActiveService]; exists {
							svc.Reset()
							svc.Blur()
							m.CurrentSvc = svc

							// Sync Window Size
							if m.Width > 0 && m.Height > 0 {
								newModel, _ := svc.Update(tea.WindowSizeMsg{
									Width:  m.Width,
									Height: m.Height,
								})
								if updatedSvc, ok := newModel.(services.Service); ok {
									svc = updatedSvc
									m.ServiceMap[m.ActiveService] = svc
									m.CurrentSvc = svc
								}
							}

							// trigger refresh
							cmds = append(cmds, func() tea.Msg { return svc.Refresh()() })
						}
						m.Focus = FocusSidebar
						m.Sidebar.Active = true
					}
					// Close Palette
					m.Focus = m.LastFocus
					// If we switched view, we might want to focus something specific?
					// For now, restore last focus (which might be weird if we changed views)
					// Actually, if we switched service, we force FocusSidebar above.
					if route.View == core.ViewHome {
						m.Focus = FocusSidebar // or menu
					}

					m.Navigation.PaletteActive = false
					m.StatusBar.Mode = "NORMAL"
					m.StatusBar.Message = "Ready"
					m.Palette.TextInput.Reset()
					m.Navigation.FilterCommands("")
				}
				return m, tea.Batch(cmds...)
			}

			// Forward other keys to Palette Input
			var cmd tea.Cmd
			m.Palette, cmd = m.Palette.Update(msg)
			cmds = append(cmds, cmd)

			// Update Suggestions
			if m.Palette.TextInput.Value() != m.Navigation.Query {
				m.Navigation.FilterCommands(m.Palette.TextInput.Value())
			}

			return m, tea.Batch(cmds...)
		}

		// HOME MODE
		if m.ViewMode == ViewHome && !m.ShowHelp && m.Focus != FocusPalette {
			switch msg.String() {
			case "enter":
				// Select service
				selected := m.HomeMenu.SelectedItem()
				if !selected.IsComing { // Only allow entering implemented services
					m.ViewMode = ViewService
					m.ActiveService = selected.ShortName

					// Sync Sidebar selection
					for i, item := range m.Sidebar.Items {
						if item.ShortName == selected.ShortName {
							m.Sidebar.Cursor = i
							break
						}
					}

					// Switch Context
					m.Focus = FocusMain // Default focus to content? Or Sidebar? Spec says sidebar default.
					m.Focus = FocusSidebar
					m.Sidebar.Active = true

					// Update Current Service
					if svc, exists := m.ServiceMap[m.ActiveService]; exists {
						svc.Reset() // Reset state (fix Bug 2)
						svc.Blur()  // Ensure dimmed state initially (Fix UX Focus)
						m.CurrentSvc = svc

						// Sync Window Size immediately implementation (Fix Bug: Truncated list on entry)
						if m.Width > 0 && m.Height > 0 {
							newModel, _ := svc.Update(tea.WindowSizeMsg{
								Width:  m.Width,
								Height: m.Height,
							})
							if updatedSvc, ok := newModel.(services.Service); ok {
								svc = updatedSvc
								m.ServiceMap[m.ActiveService] = svc
								m.CurrentSvc = svc // Update current pointer too
							}
						}

						// Trigger Refresh
						cmds = append(cmds, func() tea.Msg { return svc.Refresh()() })
					}
				}
				return m, tea.Batch(cmds...)
			}

			// Update Home Menu
			m.HomeMenu, cmd = m.HomeMenu.Update(msg)
			return m, cmd
		}

		// SERVICE MODE
		if m.ViewMode == ViewService {
			// Handle 'q' explicitly for hierarchy navigation (Fix Bug 1)
			if msg.String() == "q" && !m.ShowHelp && m.Focus != FocusPalette {
				if m.CurrentSvc != nil && m.CurrentSvc.IsRootView() {
					// If at root of service, go back to Home
					m.ViewMode = ViewHome
					m.Sidebar.Active = false
					m.HomeMenu.IsFocused = true
					return m, nil
				}
				// If not at root (e.g. detailed view), pass 'q' to service
				// to let it handle "back" logic
			}

			// Focus Switching - Handle BEFORE forwarding to service
			if m.Focus != FocusPalette && !m.ShowHelp {
				switch msg.String() {
				case "left":
					// Always allow escaping to sidebar with Left Arrow
					if m.Focus == FocusMain {
						m.Focus = FocusSidebar
						m.Sidebar.Active = true
						if m.CurrentSvc != nil {
							m.CurrentSvc.Blur() // Dim selection
						}
						// Do not forward 'left' to service
						return m, nil
					}
				case "h":
					// Only allow 'h' to switch focus if we are in sidebar
					// In FocusMain, 'h' is reserved for SSH
					if m.Focus == FocusSidebar {
						// Actually 'h' in sidebar usually collapses or does nothing contextually
						// But for now let's just keep 'left' behavior or ignore it for consistency
					}
				case "right":
					if m.Focus == FocusSidebar && m.Sidebar.Visible {
						m.Focus = FocusMain
						m.Sidebar.Active = false
						if m.CurrentSvc != nil {
							m.CurrentSvc.Focus() // Highlight selection
						}
						return m, nil
					}
				case "l":
					// 'l' in sidebar can move to main
					if m.Focus == FocusSidebar && m.Sidebar.Visible {
						m.Focus = FocusMain
						m.Sidebar.Active = false
						if m.CurrentSvc != nil {
							m.CurrentSvc.Focus() // Highlight selection
						}
						return m, nil
					}
				case "enter":
					// 'enter' in sidebar also moves to main
					if m.Focus == FocusSidebar && m.Sidebar.Visible {
						m.Focus = FocusMain
						m.Sidebar.Active = false
						if m.CurrentSvc != nil {
							m.CurrentSvc.Focus()
						}
						return m, nil
					}
				}
			}

			// If Focus is Main and we have an active service, forward keys
			if m.Focus == FocusMain && m.CurrentSvc != nil {
				var newModel tea.Model
				newModel, cmd = m.CurrentSvc.Update(msg)
				if updatedSvc, ok := newModel.(services.Service); ok {
					m.CurrentSvc = updatedSvc
					m.ServiceMap[m.CurrentSvc.ShortName()] = updatedSvc
				}
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		availableHeight := msg.Height - 1
		m.Sidebar.Height = availableHeight
		m.StatusBar.Width = msg.Width
	}

	// Global Updates
	if m.ShowHelp {
		return m, nil
	}

	// Service Mode Specific Updates
	if m.ViewMode == ViewService {
		// Update Sidebar
		m.Sidebar, cmd = m.Sidebar.Update(msg)
		cmds = append(cmds, cmd)

		// Check for Sidebar Selection Changes
		selectedSvc := m.Sidebar.SelectedService()
		if selectedSvc.ShortName != "" && m.ActiveService != selectedSvc.ShortName {
			m.ActiveService = selectedSvc.ShortName
			if svc, exists := m.ServiceMap[m.ActiveService]; exists {
				svc.Reset() // Reset state (fix Bug 2)

				// Sync Window Size immediately so table renders correctly
				if m.Width > 0 && m.Height > 0 {
					newModel, _ := svc.Update(tea.WindowSizeMsg{
						Width:  m.Width,
						Height: m.Height,
					})
					if updatedSvc, ok := newModel.(services.Service); ok {
						svc = updatedSvc
						m.ServiceMap[m.ActiveService] = svc
					}
				}

				m.CurrentSvc = svc
				// Trigger Refresh
				cmds = append(cmds, func() tea.Msg { return svc.Refresh()() })
			} else {
				m.CurrentSvc = nil
			}
		}

		// Update Active Service (if it has background work)
		if m.CurrentSvc != nil {
			var newModel tea.Model
			newModel, cmd = m.CurrentSvc.Update(msg)
			if updatedSvc, ok := newModel.(services.Service); ok {
				m.CurrentSvc = updatedSvc
				m.ServiceMap[m.CurrentSvc.ShortName()] = updatedSvc
			}
			cmds = append(cmds, cmd)
		}
	}

	// Update StatusBar
	m.StatusBar, cmd = m.StatusBar.Update(msg)
	cmds = append(cmds, cmd)

	// Dynamic Help Text
	if m.ViewMode == ViewHome {
		m.StatusBar.SetHelpText("q:Quit  ?:Help  Enter:Select")
	} else if m.CurrentSvc != nil {
		m.StatusBar.SetHelpText(m.CurrentSvc.HelpText())
	}

	return m, tea.Batch(cmds...)
}

// View renders the current UI based on state
// See home.go for the actual view logic
