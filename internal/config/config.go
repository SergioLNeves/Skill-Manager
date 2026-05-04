package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Settings holds user-configurable application settings.
type Settings struct {
	WorkspaceRoots     []string `json:"workspaceRoots"`
	GlobalSkillSources []string `json:"globalSkillSources"`
	// Legacy fields kept for backward compatibility with old settings.json files.
	SkillsHome   string   `json:"skillsHome,omitempty"`
	SkillSources []string `json:"skillSources,omitempty"`
}

func DefaultSettings(homeDir string) Settings {
	return Settings{
		WorkspaceRoots:     []string{filepath.Join(homeDir, "dev")},
		GlobalSkillSources: []string{filepath.Join(homeDir, ".claude", "skills")},
	}
}

// EffectiveGlobalSources returns the directories to scan for global skills.
// Falls back through legacy fields for backward compatibility.
func (s Settings) EffectiveGlobalSources() []string {
	if len(s.GlobalSkillSources) > 0 {
		return s.GlobalSkillSources
	}
	if len(s.SkillSources) > 0 {
		return s.SkillSources
	}
	if s.SkillsHome != "" {
		return []string{s.SkillsHome}
	}
	return nil
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
