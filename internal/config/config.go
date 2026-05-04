package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Settings holds user-configurable application settings.
type Settings struct {
	WorkspaceRoots []string `json:"workspaceRoots"`
	SkillsHome     string   `json:"skillsHome"`
}

func DefaultSettings(homeDir string) Settings {
	return Settings{
		SkillsHome:     filepath.Join(homeDir, ".skills-manager", "skills"),
		WorkspaceRoots: []string{filepath.Join(homeDir, "dev")},
	}
}

// Load reads settings from path, returning defaults on missing file.
func Load(path string, defaults Settings) (Settings, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return defaults, nil
	}
	if err != nil {
		return defaults, err
	}
	var s Settings
	if err = json.Unmarshal(data, &s); err != nil {
		return defaults, err
	}
	return s, nil
}

// Save writes settings to path (creates parent directories).
func Save(path string, s Settings) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
