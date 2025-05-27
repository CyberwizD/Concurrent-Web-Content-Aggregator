package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create temporary test files
	tmpDir := t.TempDir()

	mainConfig := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(mainConfig, []byte(`
		workers: 5
		interval: 300
		output_dir: "./output"
	`), 0644)

	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	sourcesConfig := filepath.Join(tmpDir, "sources.yaml")
	err = os.WriteFile(sourcesConfig, []byte(`
		name: "Example Source"
  		url: "https://example.com"
  		type: "rss"
	`), 0644)

	if err != nil {
		t.Fatalf("Failed to create test sources file: %v", err)
	}

	tests := []struct {
		name           string
		configPath     string
		sourcesPath    string
		wantErr        bool
		checkSourceLen int
	}{
		{
			name:           "Load both config and sources",
			configPath:     mainConfig,
			sourcesPath:    sourcesConfig,
			wantErr:        false,
			checkSourceLen: 1,
		},
		{
			name:           "Load only main config",
			configPath:     mainConfig,
			sourcesPath:    "",
			wantErr:        false,
			checkSourceLen: 0,
		},
		{
			name:           "Invalid config path",
			configPath:     "nonexistent.yaml",
			sourcesPath:    sourcesConfig,
			wantErr:        true,
			checkSourceLen: 0,
		},
		{
			name:           "Invalid sources path",
			configPath:     mainConfig,
			sourcesPath:    "nonexistent.yaml",
			wantErr:        true,
			checkSourceLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := LoadConfig(tt.configPath, tt.sourcesPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			if len(cfg.Sources.Sources) != tt.checkSourceLen {
				t.Errorf("LoadConfig() got %d sources, want %d", len(cfg.Sources.Sources), tt.checkSourceLen)
			}

			// Verify main config values
			if cfg.Workers != 5 {
				t.Errorf("LoadConfig() got Workers = %d, want 5", cfg.Workers)
			}

			if cfg.Interval != 300 {
				t.Errorf("LoadConfig() got Interval = %d, want 300", cfg.Interval)
			}

			if cfg.OutputDir != "./output" {
				t.Errorf("LoadConfig() got OutputDir = %s, want ./output", cfg.OutputDir)
			}
		})
	}
}
