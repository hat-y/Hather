package source

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hat-y/Hather/internal/config"
	"github.com/hat-y/Hather/internal/model"
)

type Source interface {
	Search(context.Context, string) ([]model.Wallpaper, error)
	Fetch(context.Context, model.Wallpaper) (model.Artifact, error)
}
type Error struct{ Class, Message string }

func (e Error) Error() string { return e.Message }
func Class(err error) string {
	if e, ok := err.(Error); ok {
		return e.Class
	}
	return "unavailable"
}

// RequestPolicy admits a Wallhaven API request without sleeping or retrying it.
type RequestPolicy func() bool

// NewRequestPolicy enforces Wallhaven's documented 45 API calls per minute limit.
func NewRequestPolicy() RequestPolicy { return newRequestPolicy(time.Now) }

func newRequestPolicy(clock func() time.Time) RequestPolicy {
	var mu sync.Mutex
	var requests []time.Time
	return func() bool {
		mu.Lock()
		defer mu.Unlock()
		now := clock()
		cutoff := now.Add(-time.Minute)
		for len(requests) > 0 && requests[0].Before(cutoff) {
			requests = requests[1:]
		}
		if len(requests) >= 45 {
			return false
		}
		requests = append(requests, now)
		return true
	}
}

type Wallhaven struct {
	BaseURL, APIKey string
	Client          *http.Client
	Store           config.Store
	RequestPolicy   RequestPolicy
}

func (s Wallhaven) Search(ctx context.Context, query string) ([]model.Wallpaper, error) {
	if err := ctx.Err(); err != nil {
		return nil, Error{"cancelled", err.Error()}
	}
	if strings.TrimSpace(query) == "" {
		return nil, Error{"invalid_input", "a search query is required"}
	}
	if s.RequestPolicy != nil && !s.RequestPolicy() {
		return nil, Error{"rate_limited", "Wallhaven request policy blocked this request; retry after one minute"}
	}
	u, _ := url.Parse(s.baseURL() + "/api/v1/search")
	params := url.Values{"q": {query}}
	if s.APIKey != "" {
		params.Set("apikey", s.APIKey)
	}
	u.RawQuery = params.Encode()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	response, err := s.client().Do(req)
	if err != nil {
		return nil, sourceError(err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusTooManyRequests {
		return nil, Error{"rate_limited", "rate limited; retry later"}
	}
	if response.StatusCode == http.StatusUnauthorized {
		return nil, Error{"authentication", "Wallhaven authentication failed"}
	}
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return nil, Error{"unavailable", "Wallhaven is unavailable"}
	}
	var result struct {
		Data []struct {
			ID         string `json:"id"`
			URL        string `json:"url"`
			Path       string `json:"path"`
			Resolution string `json:"resolution"`
			Category   string `json:"category"`
			Purity     string `json:"purity"`
			Thumbs     struct {
				Small string `json:"small"`
			} `json:"thumbs"`
		} `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&result); err != nil {
		return nil, Error{"decode", "could not decode Wallhaven response"}
	}
	if len(result.Data) == 0 {
		return nil, Error{"unavailable", "Wallhaven returned no results. Try a different query or use a local image."}
	}
	out := make([]model.Wallpaper, len(result.Data))
	for i, wall := range result.Data {
		pageURL := wall.URL
		if pageURL == "" && wall.ID != "" {
			pageURL = "https://wallhaven.cc/w/" + url.PathEscape(wall.ID)
		}
		out[i] = model.Wallpaper{SourceKind: "wallhaven", ID: wall.ID, Title: "Wallhaven " + wall.ID, PageURL: pageURL, ImageURL: wall.Path, Metadata: map[string]string{"resolution": wall.Resolution, "category": wall.Category, "purity": wall.Purity}}
		if pageURL != "" {
			out[i].Attribution = "Wallhaven: " + pageURL
		}
		if out[i].ImageURL == "" {
			out[i].ImageURL = wall.Thumbs.Small
		}
	}
	return out, nil
}

func (s Wallhaven) Fetch(ctx context.Context, wall model.Wallpaper) (model.Artifact, error) {
	if err := ctx.Err(); err != nil {
		return model.Artifact{}, Error{"cancelled", err.Error()}
	}
	if wall.ID == "" || wall.ImageURL == "" {
		return model.Artifact{}, Error{"invalid_input", "wallpaper id and image URL are required"}
	}
	for _, entry := range cacheEntries(s.Store) {
		if entry.ID == wall.ID {
			if _, err := os.Stat(entry.Path); err == nil {
				return artifact(wall.ID, entry.Path, "")
			}
		}
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, wall.ImageURL, nil)
	response, err := s.client().Do(req)
	if err != nil {
		return model.Artifact{}, Error{"download", "could not download wallpaper"}
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusTooManyRequests {
		return model.Artifact{}, Error{"rate_limited", "wallpaper download rate limited; retry later"}
	}
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return model.Artifact{}, Error{"download", "wallpaper download failed"}
	}
	kind := strings.Split(response.Header.Get("Content-Type"), ";")[0]
	bytes, err := io.ReadAll(io.LimitReader(response.Body, 25<<20))
	if err != nil || len(bytes) == 0 || !strings.HasPrefix(kind, "image/") {
		return model.Artifact{}, Error{"download", "wallpaper download is not an image"}
	}
	if s.Store.Paths.CacheDir == "" {
		return model.Artifact{}, Error{"download", "wallpaper cache is not configured"}
	}
	path, err := s.Store.CacheWallpaper(wall.ID, extension(wall.ImageURL, kind), bytes)
	if err != nil {
		return model.Artifact{}, Error{"download", "could not cache wallpaper"}
	}
	return artifact(wall.ID, path, kind)
}

func (s Wallhaven) baseURL() string {
	if s.BaseURL != "" {
		return s.BaseURL
	}
	return "https://wallhaven.cc"
}
func (s Wallhaven) client() *http.Client {
	if s.Client != nil {
		return s.Client
	}
	return http.DefaultClient
}
func cacheEntries(store config.Store) []config.CacheEntry {
	entries, _ := store.CacheEntries()
	return entries
}
func sourceError(err error) Error {
	if err == context.Canceled {
		return Error{"cancelled", err.Error()}
	}
	return Error{"unavailable", "Wallhaven is unavailable"}
}
func extension(raw, kind string) string {
	if ext := filepath.Ext(raw); ext != "" {
		return ext
	}
	if ext, _ := mime.ExtensionsByType(kind); len(ext) > 0 {
		return ext[0]
	}
	return ".img"
}
func artifact(id, path, kind string) (model.Artifact, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return model.Artifact{}, Error{"download", "could not read wallpaper"}
	}
	sum := sha256.Sum256(bytes)
	if kind == "" {
		kind = mime.TypeByExtension(filepath.Ext(path))
	}
	return model.Artifact{Path: path, SourceID: id, ContentType: kind, Size: int64(len(bytes)), SHA256: hex.EncodeToString(sum[:])}, nil
}
