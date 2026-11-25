package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadsPreferencesOfTheFormerName(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	old, err := path(oldDirName)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(old), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(old, []byte(`{"photo_artwork": true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if !Load().PhotoArtwork {
		t.Fatal("the former preferences were lost")
	}
	if err := Save(Config{}); err != nil {
		t.Fatal(err)
	}
	if Load().PhotoArtwork {
		t.Fatal("the former preferences replaced the new ones")
	}
}
