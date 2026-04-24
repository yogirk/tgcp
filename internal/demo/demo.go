// Package demo powers the --demo mode of tgcp. When Enabled is true, service
// API clients short-circuit their real network calls and return fixture data
// embedded in the binary. The fixtures live as JSON files under data/ so the
// website landing page can consume the same source of truth.
package demo

import (
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed data/*.json
var fs embed.FS

// Enabled is set to true by main.go when the user passes --demo. Service
// packages read this flag from their API layer.
var Enabled bool

// Load reads the named fixture (e.g. "gce") and unmarshals it into dst.
func Load(name string, dst any) error {
	path := "data/" + name + ".json"
	data, err := fs.ReadFile(path)
	if err != nil {
		return fmt.Errorf("demo fixture %q: %w", name, err)
	}
	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("demo fixture %q: %w", name, err)
	}
	return nil
}

// MustLoad is the panic-on-error variant for callers where failure is a bug.
func MustLoad(name string, dst any) {
	if err := Load(name, dst); err != nil {
		panic(err)
	}
}

// FixtureUser / FixtureProject are the synthetic identity strings shown in
// the UI banner when --demo is active.
const (
	FixtureUser    = "demo@tgcp.dev"
	FixtureProject = "tgcp-demo-project"
)
