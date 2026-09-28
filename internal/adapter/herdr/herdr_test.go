package herdr

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/hat-y/Hather/internal/model"
)

func TestApplyChangesOnlyOwnedThemeValues(t *testing.T) {
	path := copyFixture(t)
	result := (Adapter{ConfigPath: path, BoundaryRoot: filepath.Dir(path)}).Apply(context.Background(), testPalette())
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(got)
	if result.Status != model.AdapterApplied || result.ChangedPaths[0] != path || !strings.Contains(strings.Join(result.FollowUp, " "), "restart_required") {
		t.Fatalf("result = %#v", result)
	}
	for _, want := range []string{
		`name = "hather"`, `sidebar_bg = "#101010"`, `active_row_bg = "#202020"`,
		`selection_bg = "#abcdef"`, `panel_bg = "#101010"`, `accent = "#abcdef"`,
		`red = "#ff0000"`, `green = "#00ff00"`, `blue = "#89b4fa"`, `yellow = "#f9e2af"`,
		`default_shell = "/bin/zsh"`, `prefix = "ctrl+a"`, `command = "lazygit"`, `headless_cols = 140`,
	} {
		if !strings.Contains(text, want) {
			t.Errorf("updated config missing %q\n%s", want, text)
		}
	}
}

func TestApplyReportsAbsentConfigAsUnavailable(t *testing.T) {
	path := filepath.Join(tempDir(t), "missing.toml")
	result := (Adapter{ConfigPath: path, BoundaryRoot: filepath.Dir(path)}).Apply(context.Background(), testPalette())
	if result.Status != model.AdapterUnavailable || result.ErrorClass != "config" || !strings.Contains(strings.Join(result.Diagnostics, " "), path) || !strings.Contains(strings.Join(result.Diagnostics, " "), "HERDR_CONFIG_PATH") {
		t.Fatalf("result=%#v", result)
	}
}

func TestApplyHonorsHERDRConfigPath(t *testing.T) {
	defaultPath := copyFixture(t)
	override := filepath.Join(filepath.Dir(defaultPath), "override.toml")
	original, _ := os.ReadFile(defaultPath)
	if err := os.WriteFile(override, original, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HERDR_CONFIG_PATH", override)
	result := (Adapter{ConfigPath: defaultPath, BoundaryRoot: filepath.Dir(defaultPath)}).Apply(context.Background(), testPalette())
	unchanged, _ := os.ReadFile(defaultPath)
	changed, _ := os.ReadFile(override)
	if result.Status != model.AdapterApplied || string(unchanged) != string(original) || !strings.Contains(string(changed), `name = "hather"`) {
		t.Fatalf("result=%#v default changed=%t override=%q", result, string(unchanged) != string(original), changed)
	}
}

func TestApplyRejectsOutsideBoundaryWithoutChangingConfig(t *testing.T) {
	boundary, external := tempDir(t), tempDir(t)
	path := filepath.Join(external, "config.toml")
	original := "[theme]\nname = \"old\"\n"
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	result := (Adapter{ConfigPath: path, BoundaryRoot: boundary}).Apply(context.Background(), testPalette())
	got, err := os.ReadFile(path)
	if err != nil || result.Status != model.AdapterConflict || string(got) != original {
		t.Fatalf("result=%#v config=%q err=%v", result, got, err)
	}
}

func TestApplyRejectsSymlinkedAncestorsAndLexicalEscapes(t *testing.T) {
	for _, name := range []string{"config-path-symlink", "environment-symlink", "lexical-escape"} {
		t.Run(name, func(t *testing.T) {
			dir := tempDir(t)
			external := tempDir(t)
			path := filepath.Join(external, "config.toml")
			original := "[theme]\nname = \"old\"\n"
			if err := os.WriteFile(path, []byte(original), 0600); err != nil {
				t.Fatal(err)
			}
			link := filepath.Join(dir, "link")
			if err := os.Symlink(external, link); err != nil {
				t.Fatal(err)
			}
			unsafe := filepath.Join(link, "config.toml")
			if name == "lexical-escape" {
				if err := os.Mkdir(filepath.Join(external, "nested"), 0700); err != nil {
					t.Fatal(err)
				}
				unsafe = external + string(os.PathSeparator) + "nested" + string(os.PathSeparator) + ".." + string(os.PathSeparator) + "config.toml"
			}
			adapter := Adapter{ConfigPath: unsafe, BoundaryRoot: dir}
			if name == "environment-symlink" {
				adapter.ConfigPath = filepath.Join(dir, "unused.toml")
				t.Setenv("HERDR_CONFIG_PATH", unsafe)
			}
			result := adapter.Apply(context.Background(), testPalette())
			got, err := os.ReadFile(path)
			if err != nil || result.Status != model.AdapterConflict || string(got) != original {
				t.Fatalf("result=%#v config=%q err=%v", result, got, err)
			}
		})
	}
}

func TestApplyRejectsMalformedOrAmbiguousTOMLWithoutWriting(t *testing.T) {
	for _, original := range []string{
		"[theme\nname = \"old\"\n",
		"[theme]\nname = \"old\"\nname = \"duplicate\"\n",
		"[theme]\nname = \"\"\"multiline\nvalue\"\"\"\n",
	} {
		path := filepath.Join(tempDir(t), "config.toml")
		if err := os.WriteFile(path, []byte(original), 0600); err != nil {
			t.Fatal(err)
		}
		result := (Adapter{ConfigPath: path, BoundaryRoot: filepath.Dir(path)}).Apply(context.Background(), testPalette())
		got, _ := os.ReadFile(path)
		if result.Status != model.AdapterConflict || string(got) != original || !strings.Contains(strings.Join(result.Diagnostics, " "), "not changed") {
			t.Fatalf("result=%#v changed=%t", result, string(got) != original)
		}
	}
}

func TestApplyDoesNotRewriteAssignmentsInsideMultilineStrings(t *testing.T) {
	path := filepath.Join(tempDir(t), "config.toml")
	original := "[theme.custom]\nnotes = \"\"\"\naccent = \\\"documentation example\\\"\n\"\"\"\naccent = \"#old\"\n"
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	result := (Adapter{ConfigPath: path, BoundaryRoot: filepath.Dir(path)}).Apply(context.Background(), testPalette())
	got, _ := os.ReadFile(path)
	if result.Status != model.AdapterApplied || !strings.Contains(string(got), `accent = \"documentation example\"`) || !strings.Contains(string(got), `accent = "#abcdef"`) {
		t.Fatalf("result=%#v config=%q", result, got)
	}
}

func TestApplyRejectsMultilineArrayContainingThemeSyntaxWithoutWriting(t *testing.T) {
	for _, delimiter := range []string{`"""`, `'''`} {
		t.Run(delimiter, func(t *testing.T) {
			path := filepath.Join(tempDir(t), "config.toml")
			original := "[terminal]\nstartup = [ " + delimiter + "\n[theme.custom]\naccent = \\\"embedded text\\\"\n" + delimiter + "\n]\n"
			if _, err := toml.Decode(original, &map[string]any{}); err != nil {
				t.Fatalf("fixture must be valid TOML: %v", err)
			}
			if err := os.WriteFile(path, []byte(original), 0600); err != nil {
				t.Fatal(err)
			}

			result := (Adapter{ConfigPath: path, BoundaryRoot: filepath.Dir(path)}).Apply(context.Background(), testPalette())
			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != model.AdapterConflict || string(got) != original || !strings.Contains(strings.Join(result.Diagnostics, " "), "not changed") {
				t.Fatalf("result=%#v changed=%t", result, string(got) != original)
			}
		})
	}
}

func TestApplyAppendsMissingOwnedSections(t *testing.T) {
	path := filepath.Join(tempDir(t), "config.toml")
	original := "[terminal]\ndefault_shell = \"/bin/zsh\"\n"
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	result := (Adapter{ConfigPath: path, BoundaryRoot: filepath.Dir(path)}).Apply(context.Background(), testPalette())
	got, _ := os.ReadFile(path)
	text := string(got)
	if result.Status != model.AdapterApplied || !strings.Contains(text, "# Hather-owned theme settings\n[theme]\n") || !strings.Contains(text, "[theme.custom]\n") || !strings.HasPrefix(text, original) {
		t.Fatalf("result=%#v config=%q", result, text)
	}
	for _, want := range []string{
		`sidebar_bg = "#101010"`, `active_row_bg = "#202020"`, `selection_bg = "#abcdef"`,
		`panel_bg = "#101010"`, `accent = "#abcdef"`, `red = "#ff0000"`, `green = "#00ff00"`,
	} {
		if !strings.Contains(text, want) {
			t.Errorf("appended config missing %q\n%s", want, text)
		}
	}
	if strings.Contains(text, "blue =") || strings.Contains(text, "yellow =") {
		t.Fatalf("unverified keys were appended:\n%s", text)
	}
}

func tempDir(t *testing.T) string {
	t.Helper()
	path, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func copyFixture(t *testing.T) string {
	t.Helper()
	bytes, err := os.ReadFile("testdata/config.toml")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(tempDir(t), "config.toml")
	if err := os.WriteFile(path, bytes, 0640); err != nil {
		t.Fatal(err)
	}
	return path
}

func testPalette() model.Palette {
	return model.Palette{
		Background: "#101010", Accent: "#abcdef",
		Colors: [16]string{"#000000", "#ff0000", "#00ff00", "#ffff00", "#0000ff", "", "", "", "#202020"},
	}
}
