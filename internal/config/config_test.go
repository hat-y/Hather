package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hat-y/Hather/internal/model"
)

func TestLoadReadsOrdinarySettingsWithoutCredential(t *testing.T) {
	paths := PathsForHome(t.TempDir())
	if err := os.MkdirAll(filepath.Dir(paths.ConfigFile), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paths.ConfigFile, []byte("query = \"forest\"\nenabled_adapters = [\"herdr\"]\nwallhaven_api_key = \"config-secret\"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := Load(paths.ConfigFile)
	if err != nil {
		t.Fatal(err)
	}
	if got.Query != "forest" || len(got.EnabledAdapters) != 1 || got.EnabledAdapters[0] != "herdr" {
		t.Fatalf("config = %#v", got)
	}
	payload, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), "config-secret") {
		t.Fatalf("credential leaked into config: %s", payload)
	}
}

func TestResolveSourcePrecedence(t *testing.T) {
	for _, tc := range []struct{ explicit, env, file, want string }{
		{"flag", "env", "file", "flag"}, {"", "env", "file", "environment"}, {"", "", "file", "config"}, {"", "", "", "default"},
	} {
		if got := ResolveSource(tc.explicit, tc.env, tc.file); got != tc.want {
			t.Fatalf("source=%q want %q", got, tc.want)
		}
	}
}

func TestResolvePrecedence(t *testing.T) {
	cases := []struct{ explicit, env, file, want string }{
		{"flag", "env", "file", "flag"}, {"", "env", "file", "env"}, {"", "", "file", "file"}, {"", "", "", "default"},
	}
	for _, tc := range cases {
		if got := Resolve(tc.explicit, tc.env, tc.file, "default"); got != tc.want {
			t.Errorf("Resolve(%q, %q, %q) = %q, want %q", tc.explicit, tc.env, tc.file, got, tc.want)
		}
	}
}

func TestWallhavenAPIKeyIsTransientAndRedacted(t *testing.T) {
	secret := ResolveWallhavenAPIKey("flag-secret", "env-secret")
	if secret != "flag-secret" {
		t.Fatalf("explicit API key = %q, want flag-secret", secret)
	}

	result := model.OperationResult{Status: model.OperationComplete, Diagnostics: []string{Redact("WALLHAVEN_API_KEY="+secret, secret)}}
	store := NewStore(PathsForHome(t.TempDir()))
	if err := store.SaveLastApplied(result); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveRun(result); err != nil {
		t.Fatal(err)
	}

	configBytes, err := json.Marshal(Config{Query: "forest"})
	if err != nil {
		t.Fatal(err)
	}
	stateBytes, err := os.ReadFile(store.Paths.StateFile)
	if err != nil {
		t.Fatal(err)
	}
	runEntries, err := os.ReadDir(store.Paths.RunsDir)
	if err != nil || len(runEntries) != 1 {
		t.Fatalf("run entries = %d, %v; want one", len(runEntries), err)
	}
	runBytes, err := os.ReadFile(filepath.Join(store.Paths.RunsDir, runEntries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	for _, payload := range [][]byte{configBytes, stateBytes, runBytes} {
		if strings.Contains(string(payload), secret) || !strings.Contains(string(payload), "[REDACTED]") && string(payload) != string(configBytes) {
			t.Fatalf("secret leaked into persisted payload: %s", payload)
		}
	}
}

func TestWallhavenAPIKeyFallsBackToEnvironment(t *testing.T) {
	if got := ResolveWallhavenAPIKey("", "env-secret"); got != "env-secret" {
		t.Fatalf("environment API key = %q, want env-secret", got)
	}
}

func TestPathsUseProvidedHome(t *testing.T) {
	paths := PathsForHome("/tmp/home")
	if paths.ConfigFile != "/tmp/home/.config/hather/config.toml" || paths.CacheIndex != "/tmp/home/Library/Caches/hather/index.json" {
		t.Fatalf("unexpected paths: %#v", paths)
	}
}
