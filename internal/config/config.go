// Package config stores the user's preferences between runs.
package config

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

type Config struct {
	// PhotoArtwork shows the real album art instead of the pixelated one.
	PhotoArtwork bool `json:"photo_artwork"`
}

const (
	dirName    = "applemusictui"
	oldDirName = "sr-player" // the player's former name
)

func path(name string) (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name, "config.json"), nil
}

func read(name string) ([]byte, error) {
	p, err := path(name)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(p)
}

// Load returns the saved preferences, or the defaults if there are none.
func Load() Config {
	var c Config
	data, err := read(dirName)
	if errors.Is(err, fs.ErrNotExist) {
		data, err = read(oldDirName)
	}
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
	p, err := path(dirName)
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
