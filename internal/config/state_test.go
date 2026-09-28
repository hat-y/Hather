package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hat-y/Hather/internal/model"
)

func TestLastAppliedOnlyFollowsSuccess(t *testing.T) {
	paths := PathsForHome(t.TempDir())
	store := NewStore(paths)
	failed := model.OperationResult{Status: model.OperationPrerequisiteFailure}
	if err := store.SaveLastApplied(failed); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(paths.StateFile); !os.IsNotExist(err) {
		t.Fatalf("failed operation wrote state: %v", err)
	}
	complete := model.OperationResult{Status: model.OperationComplete, Wallpaper: model.Wallpaper{ID: "wallhaven-42"}, Palette: model.Palette{EngineVersion: "v1", Accent: "#ff0000"}}
	if err := store.SaveLastApplied(complete); err != nil {
		t.Fatal(err)
	}
	state, err := os.ReadFile(paths.StateFile)
	if err != nil || string(state) == "" || string(state) == "flag-secret" || !strings.Contains(string(state), "wallhaven-42") {
		t.Fatalf("successful state was not safely persisted: %q (%v)", state, err)
	}
}

func TestRunsPruneToDefaultLimit(t *testing.T) {
	paths := PathsForHome(t.TempDir())
	store := NewStore(paths)
	for i := 0; i < DefaultRunLimit+1; i++ {
		if err := store.SaveRun(model.OperationResult{Status: model.OperationComplete, Wallpaper: model.Wallpaper{ID: string(rune('a' + i))}}); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := os.ReadDir(paths.RunsDir)
	if err != nil || len(entries) != DefaultRunLimit {
		t.Fatalf("run entries = %d, %v; want %d", len(entries), err, DefaultRunLimit)
	}
}

func TestCachePrunesToDefaultLimit(t *testing.T) {
	paths := PathsForHome(t.TempDir())
	store := NewStore(paths)
	for i := 0; i < DefaultCacheLimit+1; i++ {
		if _, err := store.CacheWallpaper(string(rune('a'+i)), ".jpg", []byte{byte(i)}); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := store.CacheEntries()
	if err != nil || len(entries) != DefaultCacheLimit {
		t.Fatalf("cache entries = %d, %v; want %d", len(entries), err, DefaultCacheLimit)
	}
	if _, err := os.Stat(filepath.Join(paths.CacheDir, "a.jpg")); !os.IsNotExist(err) {
		t.Fatalf("oldest cache artifact survived: %v", err)
	}
	if _, err := os.Stat(filepath.Join(paths.CacheDir, "u.jpg")); err != nil {
		t.Fatalf("newest cache artifact missing: %v", err)
	}
}
