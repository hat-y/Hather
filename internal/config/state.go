package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/hat-y/Hather/internal/model"
)

const (
	DefaultCacheLimit = 20
	DefaultRunLimit   = 20
)

type Store struct{ Paths Paths }

type CacheEntry struct {
	ID       string    `json:"id"`
	Path     string    `json:"path"`
	CachedAt time.Time `json:"cached_at"`
}

type cacheIndex struct {
	Entries []CacheEntry `json:"entries"`
}

func NewStore(paths Paths) Store { return Store{Paths: paths} }

func (s Store) SaveLastApplied(result model.OperationResult) error {
	if result.Status != model.OperationComplete && result.Status != model.OperationPartialFailure {
		return nil
	}
	return s.writeJSON(s.Paths.StateFile, result)
}

func (s Store) SaveRun(result model.OperationResult) error {
	if err := s.Paths.EnsureDirectories(); err != nil {
		return err
	}
	path := filepath.Join(s.Paths.RunsDir, fmt.Sprintf("%d.json", time.Now().UnixNano()))
	if err := s.writeJSON(path, result); err != nil {
		return err
	}
	entries, err := os.ReadDir(s.Paths.RunsDir)
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries[:max(0, len(entries)-DefaultRunLimit)] {
		if err := os.Remove(filepath.Join(s.Paths.RunsDir, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func (s Store) CacheWallpaper(id, extension string, bytes []byte) (string, error) {
	if id == "" || filepath.Base(id) != id || strings.Contains(id, ".") || filepath.Ext(extension) != extension {
		return "", fmt.Errorf("invalid cache key")
	}
	if err := s.Paths.EnsureDirectories(); err != nil {
		return "", err
	}
	path := filepath.Join(s.Paths.CacheDir, id+extension)
	if err := os.WriteFile(path, bytes, 0600); err != nil {
		return "", err
	}
	entries, err := s.CacheEntries()
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	now := time.Now().UTC()
	entries = append(filterCache(entries, id), CacheEntry{ID: id, Path: path, CachedAt: now})
	sort.Slice(entries, func(i, j int) bool { return entries[i].CachedAt.Before(entries[j].CachedAt) })
	for len(entries) > DefaultCacheLimit {
		if err := os.Remove(entries[0].Path); err != nil && !os.IsNotExist(err) {
			return "", err
		}
		entries = entries[1:]
	}
	return path, s.writeJSON(s.Paths.CacheIndex, cacheIndex{Entries: entries})
}

func (s Store) CacheEntries() ([]CacheEntry, error) {
	data, err := os.ReadFile(s.Paths.CacheIndex)
	if err != nil {
		return nil, err
	}
	var index cacheIndex
	return index.Entries, json.Unmarshal(data, &index)
}

func filterCache(entries []CacheEntry, id string) []CacheEntry {
	out := entries[:0]
	for _, entry := range entries {
		if entry.ID != id {
			out = append(out, entry)
		}
	}
	return out
}

func (s Store) writeJSON(path string, value any) error {
	if err := s.Paths.EnsureDirectories(); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
