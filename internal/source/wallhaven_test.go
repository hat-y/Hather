package source

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hat-y/Hather/internal/config"
	"github.com/hat-y/Hather/internal/model"
)

func TestWallhavenSearchAndFetch(t *testing.T) {
	search, err := os.ReadFile("testdata/wallhaven_search.json")
	if err != nil {
		t.Fatal(err)
	}
	image, err := os.ReadFile("testdata/wallhaven_download_small.jpg")
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/search":
			if r.URL.Query().Get("q") != "forest & sky" {
				t.Errorf("query = %q", r.URL.RawQuery)
			}
			if r.URL.Query().Get("apikey") != "key &+?" || !strings.Contains(r.URL.RawQuery, "apikey=key+%26%2B%3F") {
				t.Errorf("apikey query = %q", r.URL.RawQuery)
			}
			if r.Header.Get("Authorization") != "" {
				t.Errorf("authorization = %q", r.Header.Get("Authorization"))
			}
			w.Write(search)
		case "/image.jpg":
			if r.URL.Query().Has("apikey") || r.Header.Get("Authorization") != "" {
				t.Errorf("image request leaked credentials: query = %q, authorization = %q", r.URL.RawQuery, r.Header.Get("Authorization"))
			}
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write(image)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	s := Wallhaven{BaseURL: server.URL, APIKey: "key &+?", Client: server.Client(), Store: config.NewStore(config.PathsForHome(t.TempDir()))}
	walls, err := s.Search(context.Background(), "forest & sky")
	if err != nil {
		t.Fatal(err)
	}
	if len(walls) == 0 || walls[0].ID != "p29ke3" || walls[0].Title != "Wallhaven p29ke3" || walls[0].Metadata["resolution"] != "1920x1080" || walls[0].Metadata["category"] != "anime" || !strings.Contains(walls[0].Attribution, "wallhaven") {
		t.Fatalf("walls = %#v", walls)
	}
	walls[0].ImageURL = server.URL + "/image.jpg"
	artifact, err := s.Fetch(context.Background(), walls[0])
	if err != nil {
		t.Fatal(err)
	}
	if artifact.Size != int64(len(image)) || artifact.ContentType != "image/jpeg" {
		t.Fatalf("artifact = %#v", artifact)
	}
}

func TestWallhavenErrorClasses(t *testing.T) {
	cases := []struct {
		name        string
		status      int
		body, class string
	}{
		{"invalid input", 0, "", "invalid_input"}, {"unavailable", 503, "", "unavailable"},
		{"rate limit", 429, `{"error":"slow down"}`, "rate_limited"}, {"auth", 401, `{"error":"Unauthorized"}`, "authentication"},
		{"malformed", 200, `{`, "decode"}, {"empty", 200, `{"data":[]}`, "unavailable"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(tc.status); w.Write([]byte(tc.body)) }))
			defer server.Close()
			_, err := (Wallhaven{BaseURL: server.URL, Client: server.Client()}).Search(context.Background(), "query")
			if tc.status == 0 {
				_, err = (Wallhaven{}).Search(context.Background(), "")
			}
			if got := Class(err); got != tc.class {
				t.Fatalf("Class(%v) = %q, want %q", err, got, tc.class)
			}
			if tc.class == "rate_limited" && !strings.Contains(err.Error(), "retry") {
				t.Fatalf("rate limit guidance = %v", err)
			}
		})
	}
}

func TestWallhavenRequestPolicyBlocksLocally(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Write([]byte(`{"data":[{"id":"wall-1","url":"https://wallhaven.cc/w/wall-1","path":"https://w.wallhaven.cc/wall-1.jpg"}]}`))
	}))
	defer server.Close()

	allowed := true
	s := Wallhaven{BaseURL: server.URL, Client: server.Client(), RequestPolicy: func() bool {
		if allowed {
			allowed = false
			return true
		}
		return false
	}}
	if _, err := s.Search(context.Background(), "forest"); err != nil {
		t.Fatalf("first request error = %v", err)
	}
	if _, err := s.Search(context.Background(), "forest"); Class(err) != "rate_limited" || !strings.Contains(err.Error(), "retry") {
		t.Fatalf("locally blocked error = %v", err)
	}
	if calls != 1 {
		t.Fatalf("HTTP requests = %d, want 1", calls)
	}
}

func TestNewRequestPolicyAllowsDocumentedMinuteBudget(t *testing.T) {
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	policy := newRequestPolicy(func() time.Time { return now })
	for i := 0; i < 45; i++ {
		if !policy() {
			t.Fatalf("request %d unexpectedly blocked", i+1)
		}
	}
	if policy() {
		t.Fatal("46th request was allowed")
	}
	now = now.Add(time.Minute + time.Nanosecond)
	if !policy() {
		t.Fatal("request after one-minute window was blocked")
	}
}

func TestWallhavenDownloadRateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(429) }))
	defer server.Close()
	_, err := (Wallhaven{Client: server.Client()}).Fetch(context.Background(), model.Wallpaper{ID: "x", ImageURL: server.URL})
	if Class(err) != "rate_limited" || !strings.Contains(err.Error(), "retry") {
		t.Fatalf("download error = %v (%s)", err, Class(err))
	}
}

func TestWallhavenEmptySearchGuidanceAndAttribution(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Has("apikey") || r.Header.Get("Authorization") != "" {
			t.Errorf("keyless request has credentials: query = %q, authorization = %q", r.URL.RawQuery, r.Header.Get("Authorization"))
		}
		if r.URL.Query().Get("q") == "empty" {
			w.Write([]byte(`{"data":[]}`))
			return
		}
		w.Write([]byte(`{"data":[{"id":"abc","path":"https://example.test/a.jpg"},{"id":"def","url":"https://wallhaven.cc/w/def","path":"https://example.test/b.jpg"}]}`))
	}))
	defer server.Close()
	s := Wallhaven{BaseURL: server.URL, Client: server.Client()}
	_, err := s.Search(context.Background(), "empty")
	if Class(err) != "unavailable" || !strings.Contains(err.Error(), "Try a different query") {
		t.Fatalf("empty error = %v", err)
	}
	walls, err := s.Search(context.Background(), "found")
	if err != nil || len(walls) != 2 {
		t.Fatalf("results = %#v, %v", walls, err)
	}
	for i, want := range []string{"https://wallhaven.cc/w/abc", "https://wallhaven.cc/w/def"} {
		if walls[i].PageURL != want || walls[i].Attribution != "Wallhaven: "+want {
			t.Fatalf("wall = %#v", walls[i])
		}
	}
}

func TestWallhavenMissingIdentityDoesNotInventAttribution(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":[{"path":"https://example.test/a.jpg"}]}`))
	}))
	defer server.Close()
	walls, err := (Wallhaven{BaseURL: server.URL, Client: server.Client()}).Search(context.Background(), "query")
	if err != nil || len(walls) != 1 || walls[0].PageURL != "" || walls[0].Attribution != "" {
		t.Fatalf("walls = %#v, err = %v", walls, err)
	}
}

func TestWallhavenDownloadAndCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "no", http.StatusBadGateway) }))
	defer server.Close()
	s := Wallhaven{Client: server.Client()}
	if _, err := s.Fetch(context.Background(), model.Wallpaper{ID: "x", ImageURL: server.URL}); Class(err) != "download" {
		t.Fatalf("download class = %q", Class(err))
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := s.Search(ctx, "query"); Class(err) != "cancelled" {
		t.Fatalf("cancel class = %q", Class(err))
	}
}

func TestLocalFetchNeedsNoSearch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wall.jpg")
	if err := os.WriteFile(path, []byte("image"), 0600); err != nil {
		t.Fatal(err)
	}
	artifact, err := (Local{}).Fetch(context.Background(), model.Wallpaper{SourceKind: "local", ImageURL: path})
	if err != nil || artifact.Path != path {
		t.Fatalf("artifact = %#v, err = %v", artifact, err)
	}
}
