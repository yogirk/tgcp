package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yogirk/tgcp/internal/config"
	"github.com/yogirk/tgcp/internal/core"
	"github.com/yogirk/tgcp/internal/ui"
	"github.com/yogirk/tgcp/internal/utils"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// 1. Parse Flags
	debug := flag.Bool("debug", false, "Enable debug logging")
	project := flag.String("project", "", "Override Google Cloud project ID")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("tgcp %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	// 2. Initialize Logger
	if *debug {
		if err := utils.InitLogger(); err != nil {
			fmt.Printf("Failed to init logger: %v\n", err)
			os.Exit(1)
		}
		defer utils.CloseLogger()
		utils.Log("Starting TGCP...")
	}

	// 3. Load Configuration
	cfg, err := config.LoadConfig()
	if err != nil && *debug {
		utils.Log("Error loading config: %v", err)
	}

	// 4. Authenticate
	// Project Priority: Flag > Config > Auto-detect
	targetProject := *project
	if targetProject == "" {
		targetProject = cfg.Project
	}

	// We do this synchronously for now for the MVP Foundation
	authState := core.Authenticate(context.Background(), targetProject)

	// 5. Create Version Info
	versionInfo := core.VersionInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	}

	// 6. Initialize UI Model
	initialModel := ui.InitialModel(authState, cfg, versionInfo)

	// 7. Start Bubbletea Program
	// WithMouseCellMotion enables mouse click support
	// Users can hold Shift to select text (standard terminal behavior)
	p := tea.NewProgram(initialModel, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
