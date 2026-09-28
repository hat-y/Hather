package vscode

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hat-y/Hather/internal/model"
	"github.com/tailscale/hujson"
)

func TestApplyMergesJSONCWithoutChangingUnrelatedSettings(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, "Library", "Application Support", "Code", "User", "settings.json")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	original := `{
  // user setting
  "editor.fontSize": 14,
  "workbench.colorCustomizations": {"terminal.ansiRed": "#keep", "editor.background": "#old",},
  "editor.tokenColorCustomizations": {"strings": "#keep",},
  "window.zoomLevel": 0,
}`
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}

	result := (Adapter{SettingsPath: path, BoundaryRoot: home}).Apply(context.Background(), palette())
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	standard, err := hujson.Standardize(append([]byte(nil), got...))
	if err != nil {
		t.Fatal(err)
	}
	var settings map[string]any
	if err := json.Unmarshal(standard, &settings); err != nil {
		t.Fatalf("rewritten settings are not JSONC: %v\n%s", err, got)
	}
	workbench := settings["workbench.colorCustomizations"].(map[string]any)
	tokens := settings["editor.tokenColorCustomizations"].(map[string]any)
	if result.Status != model.AdapterApplied || workbench["terminal.ansiRed"] != "#keep" || workbench["editor.background"] != "#101010" || workbench["editor.foreground"] != "#eeeeee" || tokens["strings"] != "#keep" || tokens["comments"] != "#777777" || settings["editor.fontSize"] != float64(14) || settings["workbench.colorTheme"] != nil || !strings.Contains(string(got), "// user setting") || !strings.Contains(strings.Join(result.FollowUp, " "), "reload") || strings.Contains(strings.Join(result.Diagnostics, " "), "live") {
		t.Fatalf("result=%#v settings=%#v\n%s", result, settings, got)
	}
}

func TestApplyAddsOnlyCustomizationSectionsWhenMissing(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, "settings.json")
	if err := os.WriteFile(path, []byte(`{"window.zoomLevel": 2}`), 0600); err != nil {
		t.Fatal(err)
	}
	result := (Adapter{SettingsPath: path, BoundaryRoot: home}).Apply(context.Background(), palette())
	got, _ := os.ReadFile(path)
	standard, err := hujson.Standardize(append([]byte(nil), got...))
	if err != nil {
		t.Fatal(err)
	}
	var settings map[string]any
	if err := json.Unmarshal(standard, &settings); err != nil {
		t.Fatal(err)
	}
	if result.Status != model.AdapterApplied || settings["window.zoomLevel"] != float64(2) || settings["workbench.colorCustomizations"] == nil || settings["editor.tokenColorCustomizations"] == nil || settings["workbench.colorTheme"] != nil {
		t.Fatalf("result=%#v settings=%#v", result, settings)
	}
	text := string(got)
	positions := []int{strings.Index(text, `"workbench.colorCustomizations"`), strings.Index(text, `"editor.background"`), strings.Index(text, `"editor.foreground"`), strings.Index(text, `"editor.lineHighlightBackground"`), strings.Index(text, `"editor.tokenColorCustomizations"`)}
	for i := 1; i < len(positions); i++ {
		if positions[i-1] < 0 || positions[i] <= positions[i-1] {
			t.Fatalf("customization order is not deterministic: %v\n%s", positions, got)
		}
	}
}

func TestApplyRejectsUnsafeSettingsWithoutWriting(t *testing.T) {
	for _, original := range []string{
		`[]`,
		`{"broken": }`,
		`{"workbench.colorCustomizations": {}, "workbench.colorCustomizations": {}}`,
		`{"workbench.colorCustomizations": {"editor.background": "#one", "editor.background": "#two"}}`,
	} {
		home := t.TempDir()
		path := filepath.Join(home, "settings.json")
		if err := os.WriteFile(path, []byte(original), 0600); err != nil {
			t.Fatal(err)
		}
		result := (Adapter{SettingsPath: path, BoundaryRoot: home}).Apply(context.Background(), palette())
		got, _ := os.ReadFile(path)
		if result.Status != model.AdapterConflict || string(got) != original || !strings.Contains(strings.Join(result.Diagnostics, " "), "not changed") {
			t.Fatalf("result=%#v changed=%t", result, string(got) != original)
		}
	}
}

func TestApplyReportsAbsentParentAndBoundaryConflict(t *testing.T) {
	home := t.TempDir()
	missing := filepath.Join(home, "missing", "settings.json")
	if result := (Adapter{SettingsPath: missing, BoundaryRoot: home}).Apply(context.Background(), palette()); result.Status != model.AdapterUnavailable {
		t.Fatalf("missing parent result=%#v", result)
	}

	outside := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(outside, []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	result := (Adapter{SettingsPath: outside, BoundaryRoot: home}).Apply(context.Background(), palette())
	got, _ := os.ReadFile(outside)
	if result.Status != model.AdapterConflict || string(got) != `{}` {
		t.Fatalf("boundary result=%#v content=%q", result, got)
	}
}

func palette() model.Palette {
	return model.Palette{Background: "#101010", Foreground: "#eeeeee", Muted: "#777777", Accent: "#abcdef"}
}
