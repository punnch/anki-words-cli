package repository

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/punnch/ankiwords/internal/model"
)

func TestFileRepositoryReturnsMissingUser(t *testing.T) {
	repo, err := NewFileRepository(filepath.Join(t.TempDir(), "settings.json"))
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	_, ok, err := repo.GetUser("1")
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if ok {
		t.Fatal("expected missing user")
	}
}

func TestFileRepositoryPersistsUser(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	repo, err := NewFileRepository(path)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	want := model.UserPrefs{
		UserID:             "1",
		ActiveDeck:         "english",
		GenerationTemplate: "template",
		PreferredModel:     "Basic",
	}
	if err := repo.UpsertUser(want); err != nil {
		t.Fatalf("upsert user: %v", err)
	}

	reloaded, err := NewFileRepository(path)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	got, ok, err := reloaded.GetUser("1")
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if !ok {
		t.Fatal("expected saved user")
	}
	if got != want {
		t.Fatalf("prefs mismatch:\n got: %#v\nwant: %#v", got, want)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat settings file: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("settings file permissions = %v, want 0600", info.Mode().Perm())
	}
}
