package config

import "path/filepath"

type Storage struct {
	// Data is the path to the root directory
	Data           string `yaml:"data"             conf:"default:./.data"`
	PocketBaseDir  string `yaml:"pocketbase-dir"   conf:""`
}

// PocketBaseDataDir returns the PocketBase data directory, defaulting to {Data}/pb_data.
func (s Storage) PocketBaseDataDir() string {
	if s.PocketBaseDir != "" {
		return s.PocketBaseDir
	}
	return filepath.Join(s.Data, "pb_data")
}
