package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Project  string         `yaml:"project"`
	Region   string         `yaml:"region"`
	Zone     string         `yaml:"zone"`
	UI       UIConfig       `yaml:"ui"`
	Features FeaturesConfig `yaml:"features"`
}

type UIConfig struct {
	SidebarVisible  bool   `yaml:"sidebar_visible"`
	RefreshInterval int    `yaml:"refresh_interval"`
	DefaultView     string `yaml:"default_view"`
}

type FeaturesConfig struct {
	EnableGCE      bool `yaml:"enable_gce"`
	EnableCloudSQL bool `yaml:"enable_cloudsql"`
}

func DefaultConfig() *Config {
	return &Config{
		UI: UIConfig{
			SidebarVisible:  true,
			RefreshInterval: 30,
			DefaultView:     "home",
		},
		Features: FeaturesConfig{
			EnableGCE:      true,
			EnableCloudSQL: true,
		},
	}
}

func LoadConfig() (*Config, error) {
	cfg := DefaultConfig()

	home, err := os.UserHomeDir()
	if err != nil {
		return cfg, err
	}

	configPath := filepath.Join(home, ".tgcprc")
	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return cfg, nil // Use defaults
	}
	if err != nil {
		return cfg, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
