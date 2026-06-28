package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/punnch/ankiwords/internal/model"
)

// FileRepository stores user preferences in a local JSON file.
type FileRepository struct {
	path string
	mu   sync.Mutex
}

type preferencesFile struct {
	Users map[string]model.UserPrefs `json:"users"`
}

// NewFileRepository creates a repository backed by path.
func NewFileRepository(path string) (*FileRepository, error) {
	if path == "" {
		return nil, fmt.Errorf("settings file path is required")
	}

	return &FileRepository{path: path}, nil
}

// GetUser loads preferences by local profile ID.
func (r *FileRepository) GetUser(userID string) (model.UserPrefs, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := r.load()
	if err != nil {
		return model.UserPrefs{}, false, err
	}

	prefs, ok := data.Users[userID]
	return prefs, ok, nil
}

// UpsertUser creates or replaces preferences for a local profile.
func (r *FileRepository) UpsertUser(prefs model.UserPrefs) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := r.load()
	if err != nil {
		return err
	}

	if data.Users == nil {
		data.Users = make(map[string]model.UserPrefs)
	}
	data.Users[prefs.UserID] = prefs

	return r.save(data)
}

func (r *FileRepository) load() (preferencesFile, error) {
	content, err := os.ReadFile(r.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return preferencesFile{Users: make(map[string]model.UserPrefs)}, nil
		}
		return preferencesFile{}, err
	}

	if len(content) == 0 {
		return preferencesFile{Users: make(map[string]model.UserPrefs)}, nil
	}

	var data preferencesFile
	if err := json.Unmarshal(content, &data); err != nil {
		return preferencesFile{}, fmt.Errorf("decode settings file: %w", err)
	}
	if data.Users == nil {
		data.Users = make(map[string]model.UserPrefs)
	}

	return data, nil
}

func (r *FileRepository) save(data preferencesFile) error {
	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')

	tmp, err := os.CreateTemp(dir, ".settings-*.json")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(content); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Chmod(tmpName, 0o600); err != nil {
		os.Remove(tmpName)
		return err
	}

	return os.Rename(tmpName, r.path)
}
