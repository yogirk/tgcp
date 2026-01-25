package ui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yogirk/tgcp/internal/config"
	"github.com/yogirk/tgcp/internal/core"
	"github.com/yogirk/tgcp/internal/services"
	"github.com/yogirk/tgcp/internal/services/bigquery"
	"github.com/yogirk/tgcp/internal/services/bigtable"
	"github.com/yogirk/tgcp/internal/services/cloudrun"
	"github.com/yogirk/tgcp/internal/services/cloudsql"
	"github.com/yogirk/tgcp/internal/services/dataflow"
	"github.com/yogirk/tgcp/internal/services/dataproc"
	"github.com/yogirk/tgcp/internal/services/disks"
	"github.com/yogirk/tgcp/internal/services/firestore"
	"github.com/yogirk/tgcp/internal/services/gce"
	"github.com/yogirk/tgcp/internal/services/gcs"
	"github.com/yogirk/tgcp/internal/services/gke"
	"github.com/yogirk/tgcp/internal/services/iam"
	"github.com/yogirk/tgcp/internal/services/logging"
	"github.com/yogirk/tgcp/internal/services/net"
	"github.com/yogirk/tgcp/internal/services/overview"
	"github.com/yogirk/tgcp/internal/services/pubsub"
	"github.com/yogirk/tgcp/internal/services/redis"
	"github.com/yogirk/tgcp/internal/services/secrets"
	"github.com/yogirk/tgcp/internal/services/spanner"
	"github.com/yogirk/tgcp/internal/ui/components"
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
	Palette   components.PaletteModel  // Added
	Toast     *components.ToastModel   // Toast notification (nil when hidden)
	Spinner   components.SpinnerModel  // Global loading spinner

	// State
	ViewMode      ViewMode // Added
	Focus         FocusArea
	LastFocus     FocusArea
	ShowHelp      bool
	ActiveService string

	// External Managers
	ProjectManager  *core.ProjectManager
	ServiceRegistry *core.ServiceRegistry
}

// InitialModel returns the initial state of the application
func InitialModel(authState core.AuthState, cfg *config.Config) MainModel {
	// Initialize Cache
	cache := core.NewCache()

	// Create service registry and register all services
	registry := core.NewServiceRegistry(cache)
	registerAllServices(registry)

	// Create service map but don't initialize services yet (lazy initialization)
	// Services will be initialized on first access
	svcMap := registry.InitializeAll(context.Background(), authState.ProjectID)

	// Initialize Components
	sb := components.NewSidebar()
	sb.Visible = cfg.UI.SidebarVisible
	statusBar := components.NewStatusBar()
	statusBar.SetFocusPane("HOME")

	return MainModel{
		AuthState:       authState,
		Navigation:      core.NewNavigation(),
		Sidebar:         sb,
		HomeMenu:        components.NewHomeMenu(),
		StatusBar:       statusBar,
		Palette:         components.NewPalette(),
		Spinner:         components.NewSpinner(),
		Focus:           FocusSidebar,
		ViewMode:        ViewHome,
		ServiceMap:      svcMap,
		ProjectManager:  core.NewProjectManager(cache),
		ServiceRegistry: registry,
	}
}

// Init initializes the bubbletea program
func (m MainModel) Init() tea.Cmd {
	return tea.EnableMouseCellMotion
}

// getOrInitializeService gets a service from the map, initializing it lazily if needed
// This implements lazy initialization - services are only initialized when first accessed
func (m *MainModel) getOrInitializeService(ctx context.Context, serviceName string) (services.Service, error) {
	// First check if service exists in map
	if svc, exists := m.ServiceMap[serviceName]; exists {
		// Service exists, but may not be initialized yet
		// Use registry to ensure it's initialized
		if m.ServiceRegistry != nil {
			initializedSvc, err := m.ServiceRegistry.GetOrInitializeService(ctx, serviceName)
			if err != nil {
				return svc, err // Return original service if init fails
			}
			if initializedSvc != nil {
				// Update the map with the initialized service
				m.ServiceMap[serviceName] = initializedSvc
				return initializedSvc, nil
			}
		}
		return svc, nil
	}
	
	// Service doesn't exist in map - try to get it from registry (lazy creation)
	if m.ServiceRegistry != nil {
		svc, err := m.ServiceRegistry.GetOrInitializeService(ctx, serviceName)
		if err != nil {
			return nil, err
		}
		if svc != nil {
			// Add to map
			m.ServiceMap[serviceName] = svc
			return svc, nil
		}
	}
	
	return nil, nil // Service not found
}

// Update handles messages and updates the model
func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	defer m.syncStatusBarFocus()

	switch msg := msg.(type) {
	// Toast Notifications
	case core.ToastMsg:
		m.Toast = components.NewToastFromMsg(msg)
		return m, m.Toast.DismissCmd()

	case components.ToastDismissMsg:
		m.Toast = nil
		return m, nil

	// Loading Spinner
	case core.LoadingMsg:
		if msg.IsLoading {
			cmd = m.Spinner.Start(msg.Message)
			return m, cmd
		} else {
			m.Spinner.Stop()
			return m, nil
		}

	case components.SpinnerTickMsg:
		var cmds []tea.Cmd
		// Update main model spinner
		m.Spinner, cmd = m.Spinner.Update(msg)
		cmds = append(cmds, cmd)
		// Also forward to current service for its own spinner animation
		if m.ViewMode == ViewService && m.CurrentSvc != nil {
			var newModel tea.Model
			newModel, cmd = m.CurrentSvc.Update(msg)
			if updatedSvc, ok := newModel.(services.Service); ok {
				m.CurrentSvc = updatedSvc
				m.ServiceMap[m.CurrentSvc.ShortName()] = updatedSvc
			}
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

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
			case "q", "esc":
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
				m.setFocus(FocusPalette)
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
						m.setFocus(FocusMain)
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
				m.setFocus(m.LastFocus)
				m.Navigation.PaletteActive = false
				m.StatusBar.Message = "Ready"
				m.Palette.TextInput.Reset()        // Clear input
				m.Navigation.RestoreBaseCommands() // Reset to default commands
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
						// Check for Project Switch
						if len(route.ID) > 15 && route.ID[:15] == "SWITCH_PROJECT:" {
							newProjectID := route.ID[15:]
							m.AuthState.ProjectID = newProjectID

							// Re-initialize all services with new project using registry
							if m.ServiceRegistry != nil {
								m.ServiceRegistry.ReinitializeAll(context.Background(), newProjectID, m.ServiceMap)
							}

							m.StatusBar.Message = "Switched to project: " + newProjectID
							m.Navigation.RestoreBaseCommands()
							m.ViewMode = ViewHome
							m.Sidebar.Active = false
						} else {
							m.ViewMode = ViewHome
							m.Sidebar.Active = false
						}
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
						// Get or initialize service lazily
						svc, err := m.getOrInitializeService(context.Background(), m.ActiveService)
						if err != nil {
							cmds = append(cmds, func() tea.Msg {
								return core.StatusMsg{Message: "Failed to initialize service: " + err.Error(), IsError: true}
							})
						} else if svc != nil {
							svc.Reset()
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
						m.setFocus(FocusMain)
						m.Sidebar.Active = false
					} else if route.View == core.ViewProjectSwitcher {
						// Trigger fetch projects
						cmds = append(cmds, func() tea.Msg {
							projects, err := m.ProjectManager.ListProjects(context.Background())
							if err != nil {
								return core.StatusMsg{Message: "Failed to list projects: " + err.Error(), IsError: true}
							}
							return projects
						})
						// Keep palette open? Yes.
						// Status update?
						m.StatusBar.Message = "Fetching projects..."
						m.setFocus(FocusPalette)
						return m, tea.Batch(cmds...)
					}
					// Close Palette
					m.setFocus(m.LastFocus)
					// If we switched view, we might want to focus something specific?
					// For now, restore last focus (which might be weird if we changed views)
					// Actually, if we switched service, we force FocusSidebar above.
					if route.View == core.ViewHome {
						m.setFocus(FocusSidebar) // or menu
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
			case "enter", " ":
				// If on category header, toggle it
				if m.HomeMenu.IsOnCategory() {
					m.HomeMenu.ToggleCurrentCategory()
					return m, nil
				}
				// Select service
				selected := m.HomeMenu.SelectedItem()
				if selected.ShortName != "" && !selected.IsComing { // Only allow entering implemented services
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
					m.Sidebar.Active = false

					// Get or initialize service lazily
					svc, err := m.getOrInitializeService(context.Background(), m.ActiveService)
					if err != nil {
						cmds = append(cmds, func() tea.Msg {
							return core.StatusMsg{Message: "Failed to initialize service: " + err.Error(), IsError: true}
						})
					} else if svc != nil {
						svc.Reset() // Reset state (fix Bug 2)
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
					m.setFocus(FocusMain)
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
					m.setFocus(FocusSidebar)
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
						m.setFocus(FocusSidebar)
						m.Sidebar.Active = true
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
				case "right", "l":
					if m.Focus == FocusSidebar && m.Sidebar.Visible {
						m.setFocus(FocusMain)
						m.Sidebar.Active = false
						return m, nil
					}
				case "enter":
					// 'enter' in sidebar also moves to main
					if m.Focus == FocusSidebar && m.Sidebar.Visible {
						m.setFocus(FocusMain)
						m.Sidebar.Active = false
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

	case []core.Project:
		// Projects Fetched
		var cmds []core.Command
		for _, p := range msg {
			// Capture variable
			p := p
			cmds = append(cmds, core.Command{
				Name:        p.ID,
				Description: p.Name,
				Action: func() core.Route {
					return core.Route{
						View: core.ViewHome,
						ID:   "SWITCH_PROJECT:" + p.ID,
					}
				},
			})
		}
		m.Navigation.SetCommands(cmds)
		m.StatusBar.Message = "Select a project to switch..."
		return m, nil

	case core.SwitchToLogsMsg:
		// Switch to Logging Service
		m.ViewMode = ViewService
		m.ActiveService = "logs"
		// Sync Sidebar
		for i, item := range m.Sidebar.Items {
			if item.ShortName == "logs" {
				m.Sidebar.Cursor = i
				break
			}
		}

		svc, err := m.getOrInitializeService(context.Background(), "logs")
		if err != nil {
			cmds = append(cmds, func() tea.Msg {
				return core.StatusMsg{Message: "Failed to initialize logging service: " + err.Error(), IsError: true}
			})
			// Still allow switching? m.CurrentSvc would be nil or stale.
			// Better to alert user and not switch context if fatal?
			// But getOrInitializeService returns original svc if it existed.
		}

		if svc != nil {
			m.CurrentSvc = svc
			m.ServiceMap["logs"] = svc // Ensure map is up to date

			svc.Reset()
			// Cast to Logging Service to set filter
			// We need a way to pass filter. Is it exposed?
			// The interface Service doesn't have SetFilter.
			// We can use type assertion.
			if logSvc, ok := svc.(interface{ SetFilter(string) }); ok {
				logSvc.SetFilter(msg.Filter)
			}
			if logSvc, ok := svc.(interface{ SetReturnTo(string) }); ok {
				logSvc.SetReturnTo(msg.Source)
			}
			if logSvc, ok := svc.(interface{ SetHeading(string) }); ok {
				logSvc.SetHeading(msg.Heading)
			}

			svc.Focus()

			// Sync Window Size
			if m.Width > 0 && m.Height > 0 {
				availWidth := m.Width
				// We force sidebar visible below, so account for it now
				availWidth -= m.Sidebar.Width

				newModel, _ := svc.Update(tea.WindowSizeMsg{
					Width:  availWidth,
					Height: m.Height,
				})
				if updatedSvc, ok := newModel.(services.Service); ok {
					svc = updatedSvc
					m.ServiceMap["logs"] = svc
					m.CurrentSvc = svc
				}
			}

			// Trigger Refresh
			cmds = append(cmds, func() tea.Msg { return svc.Refresh()() })
		}
		m.Focus = FocusMain
		m.Sidebar.Active = false
		m.Sidebar.Visible = true
		return m, tea.Batch(cmds...)

	case core.SwitchToServiceMsg:
		// Switch to a specific service
		m.ViewMode = ViewService
		m.ActiveService = msg.Service
		// Sync Sidebar
		for i, item := range m.Sidebar.Items {
			if item.ShortName == msg.Service {
				m.Sidebar.Cursor = i
				break
			}
		}

		if svc, exists := m.ServiceMap[msg.Service]; exists {
			svc.Reset()
			svc.Blur() // Focus sidebar initially? or Focus service?
			// If returning from logs, maybe focus service directly?
			// Let's stick to standard flow: Focus Sidebar active.
			// But if user pressed Esc in logs, they expect to be back in the list, possibly focused on list?
			// For now, consistent behavior: Sidebar active.
			m.CurrentSvc = svc

			// Sync Window Size
			if m.Width > 0 && m.Height > 0 {
				availWidth := m.Width
				if m.Sidebar.Visible {
					availWidth -= m.Sidebar.Width
				}
				newModel, _ := svc.Update(tea.WindowSizeMsg{
					Width:  availWidth, // Use available width
					Height: m.Height,
				})
				if updatedSvc, ok := newModel.(services.Service); ok {
					svc = updatedSvc
					m.ServiceMap[msg.Service] = svc
					m.CurrentSvc = svc
				}
			}

			// Trigger Refresh?
			cmds = append(cmds, func() tea.Msg { return svc.Refresh()() })
		}
		m.Focus = FocusSidebar
		m.Sidebar.Active = true
		return m, tea.Batch(cmds...)

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
		if m.Sidebar.Active {
			selectedSvc := m.Sidebar.SelectedService()
			if selectedSvc.ShortName != "" && m.ActiveService != selectedSvc.ShortName {
				m.ActiveService = selectedSvc.ShortName
				// Get or initialize service lazily
				svc, err := m.getOrInitializeService(context.Background(), m.ActiveService)
				if err != nil {
					cmds = append(cmds, func() tea.Msg {
						return core.StatusMsg{Message: "Failed to initialize service: " + err.Error(), IsError: true}
					})
					m.CurrentSvc = nil
				} else if svc != nil {
					svc.Reset() // Reset state (fix Bug 2)

					// Sync Window Size immediately so table renders correctly
					if m.Width > 0 && m.Height > 0 {
						availWidth := m.Width
						if m.Sidebar.Visible {
							availWidth -= m.Sidebar.Width
						}
						newModel, _ := svc.Update(tea.WindowSizeMsg{
							Width:  availWidth,
							Height: m.Height,
						})
						if updatedSvc, ok := newModel.(services.Service); ok {
							svc = updatedSvc
							m.ServiceMap[m.ActiveService] = svc
						}
					}

					m.CurrentSvc = svc
					m.setFocus(m.Focus)
					// Trigger Refresh
					cmds = append(cmds, func() tea.Msg { return svc.Refresh()() })
				} else {
					m.CurrentSvc = nil
				}
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
	if m.Focus == FocusPalette {
		m.StatusBar.SetHelpText("Esc:Cancel  Enter:Run  ↑/↓:Select")
	} else if m.ViewMode == ViewHome {
		// Add help hint if help is not currently shown
		helpText := "q:Quit  Enter:Select"
		if !m.ShowHelp {
			helpText += "  ?:Help"
		}
		m.StatusBar.SetHelpText(helpText)
	} else if m.CurrentSvc != nil {
		// Get service help text and append help hint if help is not currently shown
		helpText := m.CurrentSvc.HelpText()
		if !m.ShowHelp {
			// Append help hint to service help text
			if helpText != "" {
				helpText += "  ?:Help"
			} else {
				helpText = "?:Help"
			}
		}
		m.StatusBar.SetHelpText(helpText)
	} else {
		// Fallback: show help hint if available
		if !m.ShowHelp {
			m.StatusBar.SetHelpText("?:Help")
		} else {
			m.StatusBar.SetHelpText("")
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the current UI based on state
// See home.go for the actual view logic

// registerAllServices registers all available services with the registry
// This is kept in the ui package to avoid import cycles (services import core, core shouldn't import services)
func registerAllServices(registry *core.ServiceRegistry) {
	registry.Register("overview", func(cache *core.Cache) services.Service {
		return overview.NewService(cache)
	})
	registry.Register("gce", func(cache *core.Cache) services.Service {
		return gce.NewService(cache)
	})
	registry.Register("gke", func(cache *core.Cache) services.Service {
		return gke.NewService(cache)
	})
	registry.Register("disks", func(cache *core.Cache) services.Service {
		return disks.NewService(cache)
	})
	registry.Register("pubsub", func(cache *core.Cache) services.Service {
		return pubsub.NewService(cache)
	})
	registry.Register("redis", func(cache *core.Cache) services.Service {
		return redis.NewService(cache)
	})
	registry.Register("spanner", func(cache *core.Cache) services.Service {
		return spanner.NewService(cache)
	})
	registry.Register("bigtable", func(cache *core.Cache) services.Service {
		return bigtable.NewService(cache)
	})
	registry.Register("dataflow", func(cache *core.Cache) services.Service {
		return dataflow.NewService(cache)
	})
	registry.Register("dataproc", func(cache *core.Cache) services.Service {
		return dataproc.NewService(cache)
	})
	registry.Register("firestore", func(cache *core.Cache) services.Service {
		return firestore.NewService(cache)
	})
	registry.Register("sql", func(cache *core.Cache) services.Service {
		return cloudsql.NewService(cache)
	})
	registry.Register("iam", func(cache *core.Cache) services.Service {
		return iam.NewService(cache)
	})
	registry.Register("run", func(cache *core.Cache) services.Service {
		return cloudrun.NewService(cache)
	})
	registry.Register("gcs", func(cache *core.Cache) services.Service {
		return gcs.NewService(cache)
	})
	registry.Register("bq", func(cache *core.Cache) services.Service {
		return bigquery.NewService(cache)
	})
	registry.Register("net", func(cache *core.Cache) services.Service {
		return net.NewService(cache)
	})
	registry.Register("logs", func(cache *core.Cache) services.Service {
		return logging.NewService(cache)
	})
	registry.Register("secrets", func(cache *core.Cache) services.Service {
		return secrets.NewService(cache)
	})
}

func (m *MainModel) setFocus(area FocusArea) {
	m.Focus = area
	if m.ViewMode == ViewService && m.CurrentSvc != nil {
		switch area {
		case FocusMain:
			m.CurrentSvc.Focus()
		case FocusSidebar:
			m.CurrentSvc.Blur()
		}
	}
	m.syncStatusBarFocus()
}

func (m *MainModel) syncStatusBarFocus() {
	if m.ViewMode == ViewHome {
		m.StatusBar.SetFocusPane("HOME")
		return
	}

	switch m.Focus {
	case FocusSidebar:
		m.StatusBar.SetFocusPane("SIDEBAR")
	case FocusMain:
		m.StatusBar.SetFocusPane("MAIN")
	default:
		m.StatusBar.SetFocusPane("")
	}
}
