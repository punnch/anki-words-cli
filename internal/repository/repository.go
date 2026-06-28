// Package repository provides persistence for local card creation preferences.
package repository

import "github.com/punnch/ankiwords/internal/model"

// Repository is the persistence contract required by the command handlers.
type Repository interface {
	// GetUser loads preferences by local profile ID.
	GetUser(userID string) (model.UserPrefs, bool, error)
	// UpsertUser creates or replaces preferences for a local profile.
	UpsertUser(prefs model.UserPrefs) error
}
