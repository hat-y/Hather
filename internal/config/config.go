package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Paths struct {
	ConfigFile, StateFile, RunsDir, CacheDir, CacheIndex string
}

func PathsForHome(home string) Paths {
	configDir := filepath.Join(home, ".config", "hather")
	stateDir := filepath.Join(home, ".local", "state", "hather")
	cacheDir := filepath.Join(home, "Library", "Caches", "hather")
	return Paths{filepath.Join(configDir, "config.toml"), filepath.Join(stateDir, "last-applied.json"), filepath.Join(stateDir, "runs"), cacheDir, filepath.Join(cacheDir, "index.json")}
}

func (p Paths) EnsureDirectories() error {
	for _, dir := range []string{filepath.Dir(p.ConfigFile), filepath.Dir(p.StateFile), p.RunsDir, p.CacheDir} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
		if err := os.Chmod(dir, 0700); err != nil {
			return err
		}
	}
	return nil
}

type Config struct {
	Query           string   `json:"query" toml:"query"`
	EnabledAdapters []string `json:"enabled_adapters" toml:"enabled_adapters"`
}

func Load(path string) (Config, error) {
	var config Config
	if _, err := toml.DecodeFile(path, &config); err != nil && !os.IsNotExist(err) {
		return Config{}, err
	}
	return config, nil
}

func Resolve(explicit, environment, file, fallback string) string {
	for _, value := range []string{explicit, environment, file, fallback} {
		if value != "" {
			return value
		}
	}
	return ""
}

func ResolveSource(explicit, environment, file string) string {
	switch {
	case explicit != "":
		return "flag"
	case environment != "":
		return "environment"
	case file != "":
		return "config"
	default:
		return "default"
	}
}

func ResolveWallhavenAPIKey(explicit, environment string) string {
	return Resolve(explicit, environment, "", "")
}

func ResolveAdapters(explicit []string, environment string, file, fallback []string) []string {
	if len(explicit) > 0 {
		return append([]string(nil), explicit...)
	}
	if environment != "" {
		return strings.Split(environment, ",")
	}
	if len(file) > 0 {
		return append([]string(nil), file...)
	}
	return append([]string(nil), fallback...)
}

func Redact(text string, secrets ...string) string {
	for _, secret := range secrets {
		if secret != "" {
			text = strings.ReplaceAll(text, secret, "[REDACTED]")
		}
	}
	return text
}
