// Package config stores the user's preferences between runs.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	// PhotoArtwork shows the real album art instead of the pixelated one.
	PhotoArtwork bool `json:"photo_artwork"`
}

func path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "sr-player", "config.json"), nil
}

// Load returns the saved preferences, or the defaults if there are none.
func Load() Config {
	var c Config
	p, err := path()
	if err != nil {
		return c
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return c
	}
	if json.Unmarshal(data, &c) != nil {
		return Config{}
	}
	return c
}

// Save writes the preferences.
func Save(c Config) error {
	p, err := path()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, append(data, '\n'), 0o644)
}
