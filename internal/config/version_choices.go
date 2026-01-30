package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// VersionChoices stores user choices for prompt file upgrades per application version
type VersionChoices struct {
	Choices map[string]string `yaml:"choices"` // version -> "overwrite" or "keep"
}

// GetVersionChoicesFilePath returns the full path to the version choices file
func GetVersionChoicesFilePath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "prompt-version-choices.yaml"), nil
}

// LoadVersionChoices loads version choices from file
// Returns empty choices map if file doesn't exist or is corrupted (non-fatal)
func LoadVersionChoices() (*VersionChoices, error) {
	filePath, err := GetVersionChoicesFilePath()
	if err != nil {
		return &VersionChoices{Choices: make(map[string]string)}, nil
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &VersionChoices{Choices: make(map[string]string)}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		// Non-fatal: return empty choices
		return &VersionChoices{Choices: make(map[string]string)}, nil
	}

	var choices VersionChoices
	if err := yaml.Unmarshal(data, &choices); err != nil {
		// Non-fatal: return empty choices
		return &VersionChoices{Choices: make(map[string]string)}, nil
	}

	if choices.Choices == nil {
		choices.Choices = make(map[string]string)
	}

	return &choices, nil
}

// SaveVersionChoices saves version choices to file
func SaveVersionChoices(choices *VersionChoices) error {
	// Ensure config directory exists
	if err := EnsureDirectoryStructure(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	filePath, err := GetVersionChoicesFilePath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(choices)
	if err != nil {
		return fmt.Errorf("failed to marshal version choices: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write version choices file: %w", err)
	}

	return nil
}

// GetChoiceForVersion returns the user's choice for the given application version
// Returns empty string if no choice exists for this version
func GetChoiceForVersion(version string) (string, error) {
	choices, err := LoadVersionChoices()
	if err != nil {
		return "", err
	}

	if choice, exists := choices.Choices[version]; exists {
		return choice, nil
	}

	return "", nil
}

// SetChoiceForVersion stores the user's choice for the given application version
func SetChoiceForVersion(version, choice string) error {
	choices, err := LoadVersionChoices()
	if err != nil {
		return err
	}

	if choices.Choices == nil {
		choices.Choices = make(map[string]string)
	}

	choices.Choices[version] = choice

	return SaveVersionChoices(choices)
}
