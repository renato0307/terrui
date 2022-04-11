package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Organization           string `json:"organization"`
	Workspace              string `json:"workspace"`
	WorkspaceShowVariables bool   `json:"workspace_show_vars"`

	RunID string
}

const configDirectory = ".config/terrui"
const configFile = "terrui.json"

func NewConfig() (*Config, error) {
	c := &Config{
		WorkspaceShowVariables: true,
	}
	err := c.Load()

	if errors.Is(err, os.ErrNotExist) {
		err := c.Save()
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Config) Load() error {
	configFile, err := buildConfigFilePath()
	if err != nil {
		return fmt.Errorf("invalid config file path: %w", err)
	}

	content, err := os.ReadFile(configFile)
	if errors.Is(err, os.ErrNotExist) {
		return err
	} else if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	if err = json.Unmarshal(content, c); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	return nil
}

func (c *Config) Save() error {
	configFile, err := buildConfigFilePath()
	if err != nil {
		return fmt.Errorf("invalid config file path: %w", err)
	}

	bytes, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("error saving config file: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(configFile), 0770); err != nil {
		return fmt.Errorf("error saving config file: %w", err)
	}

	err = os.WriteFile(configFile, bytes, 0733)
	if err != nil {
		return fmt.Errorf("error saving config file: %w", err)
	}

	return nil
}

func buildConfigFilePath() (string, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", userHome, configDirectory, configFile), nil
}
