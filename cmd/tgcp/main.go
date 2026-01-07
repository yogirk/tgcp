package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rk/tgcp/internal/config"
	"github.com/rk/tgcp/internal/core"
	"github.com/rk/tgcp/internal/ui"
	"github.com/rk/tgcp/internal/utils"
)

func main() {
	// 1. Parse Flags
	debug := flag.Bool("debug", false, "Enable debug logging")
	project := flag.String("project", "", "Override Google Cloud project ID")
	flag.Parse()

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

	// 5. Initialize UI Model
	initialModel := ui.InitialModel(authState, cfg)

	// 6. Start Bubbletea Program
	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
