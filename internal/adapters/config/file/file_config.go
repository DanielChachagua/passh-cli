package file

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Config holds the configuration structure saved in config.json.
type Config struct {
	Token string `json:"token"`
}

// FileConfigStorage implements ports.ConfigStorage.
type FileConfigStorage struct {
	filePath string
}

// NewFileConfigStorage initializes and returns a FileConfigStorage instance, creating the target directory if needed.
func NewFileConfigStorage() (*FileConfigStorage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(homeDir, ".ssh-manager")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	filePath := filepath.Join(dir, "config.json")
	return &FileConfigStorage{filePath: filePath}, nil
}

// SaveToken serializes the token into config.json.
func (s *FileConfigStorage) SaveToken(token string) error {
	cfg := Config{Token: token}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0600)
}

// LoadToken reads the token from config.json.
func (s *FileConfigStorage) LoadToken() (string, error) {
	if _, err := os.Stat(s.filePath); errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return "", err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", err
	}
	return cfg.Token, nil
}

// ClearToken deletes the config.json file to clear credentials.
func (s *FileConfigStorage) ClearToken() error {
	if _, err := os.Stat(s.filePath); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return os.Remove(s.filePath)
}
