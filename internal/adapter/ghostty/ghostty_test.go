package ghostty

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hat-y/Hather/internal/model"
)

func TestApplyWritesOnlyOwnedThemeAndMarkedSelection(t *testing.T) {
	dir := tempDir(t)
	config := filepath.Join(dir, "config")
	original := "font-size = 14\n# Hather-managed theme selection\ntheme = old\n"
	if err := os.WriteFile(config, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	result := Adapter{ConfigDir: dir, BoundaryRoot: dir}.Apply(context.Background(), palette())
	bytes, err := os.ReadFile(config)
	if err != nil {
		t.Fatal(err)
	}
	theme, err := os.ReadFile(filepath.Join(dir, "themes", "hather.conf"))
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != model.AdapterApplied || !strings.Contains(string(bytes), "font-size = 14\n# Hather-managed theme selection\ntheme = hather\n") || !strings.Contains(string(theme), "# Hather-owned Ghostty theme\n") {
		t.Fatalf("result=%#v config=%q theme=%q", result, bytes, theme)
	}
}

func TestApplyProtectsExistingUnownedThemeArtifact(t *testing.T) {
	for _, selection := range []string{"theme = hather\n", marker + "\ntheme = old\n"} {
		t.Run(strings.TrimSpace(selection), func(t *testing.T) {
			dir := tempDir(t)
			config := filepath.Join(dir, "config")
			original := []byte("font-size = 14\n" + selection)
			if err := os.WriteFile(config, original, 0600); err != nil {
				t.Fatal(err)
			}
			themes := filepath.Join(dir, "themes")
			if err := os.Mkdir(themes, 0700); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(themes, "hather.conf")
			userTheme := []byte("# my colors\nbackground = #abcdef\n")
			if err := os.WriteFile(path, userTheme, 0600); err != nil {
				t.Fatal(err)
			}
			got := Adapter{ConfigDir: dir, BoundaryRoot: dir}.Apply(context.Background(), palette())
			configAfter, _ := os.ReadFile(config)
			themeAfter, _ := os.ReadFile(path)
			if got.Status != model.AdapterConflict || string(configAfter) != string(original) || string(themeAfter) != string(userTheme) {
				t.Fatalf("result=%#v config=%q theme=%q", got, configAfter, themeAfter)
			}
		})
	}
}

func TestApplyUpdatesOwnedThemeDeterministically(t *testing.T) {
	dir := tempDir(t)
	config := filepath.Join(dir, "config")
	original := []byte(marker + "\ntheme = hather\nfont-size = 14\n")
	if err := os.WriteFile(config, original, 0600); err != nil {
		t.Fatal(err)
	}
	themeDir := filepath.Join(dir, "themes")
	if err := os.Mkdir(themeDir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(themeDir, "hather.conf")
	if err := os.WriteFile(path, []byte(themeHeader+"background = #000000\n"), 0600); err != nil {
		t.Fatal(err)
	}
	adapter := Adapter{ConfigDir: dir, BoundaryRoot: dir}
	for i := 0; i < 2; i++ {
		result := adapter.Apply(context.Background(), palette())
		configAfter, _ := os.ReadFile(config)
		themeAfter, _ := os.ReadFile(path)
		if result.Status != model.AdapterApplied || string(configAfter) != string(original) || string(themeAfter) != string(themeFile(palette())) {
			t.Fatalf("iteration %d: result=%#v config=%q theme=%q", i, result, configAfter, themeAfter)
		}
	}
}

func TestApplyDoesNotTreatOtherKeysAsMarkedSelection(t *testing.T) {
	dir := tempDir(t)
	config := filepath.Join(dir, "config")
	original := marker + "\ntheme-variant = custom\n"
	if err := os.WriteFile(config, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	result := Adapter{ConfigDir: dir, BoundaryRoot: dir}.Apply(context.Background(), palette())
	got, err := os.ReadFile(config)
	if err != nil || result.Status != model.AdapterConflict || string(got) != original {
		t.Fatalf("result=%#v config=%q err=%v", result, got, err)
	}
}

func TestApplyRejectsSymlinkedConfigBeforeWritingTheme(t *testing.T) {
	dir := tempDir(t)
	target := filepath.Join(dir, "real-config")
	original := marker + "\ntheme = old\n"
	if err := os.WriteFile(target, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(dir, "config")); err != nil {
		t.Fatal(err)
	}
	result := Adapter{ConfigDir: dir, BoundaryRoot: dir}.Apply(context.Background(), palette())
	got, _ := os.ReadFile(target)
	if result.Status != model.AdapterConflict || string(got) != original {
		t.Fatalf("result=%#v target=%q, want unchanged conflict", result, got)
	}
	if _, err := os.Stat(filepath.Join(dir, "themes", "hather.conf")); !os.IsNotExist(err) {
		t.Fatalf("theme artifact created before config rejection: %v", err)
	}
}

func TestApplyRejectsMalformedOrDuplicateMarkedSelection(t *testing.T) {
	for _, original := range []string{
		marker + "\ntheme\n",
		marker + "\ntheme = old\n" + marker + "\ntheme = duplicate\n",
	} {
		dir := tempDir(t)
		config := filepath.Join(dir, "config")
		if err := os.WriteFile(config, []byte(original), 0600); err != nil {
			t.Fatal(err)
		}
		result := Adapter{ConfigDir: dir, BoundaryRoot: dir}.Apply(context.Background(), palette())
		got, _ := os.ReadFile(config)
		if result.Status != model.AdapterConflict || string(got) != original {
			t.Fatalf("result=%#v config=%q, want unchanged conflict", result, got)
		}
		if _, err := os.Stat(filepath.Join(dir, "themes", "hather.conf")); !os.IsNotExist(err) {
			t.Fatalf("theme artifact created before validation: %v", err)
		}
	}
}

func TestApplyRejectsOutsideBoundaryBeforeWriting(t *testing.T) {
	boundary, external := tempDir(t), tempDir(t)
	original := marker + "\ntheme = old\n"
	if err := os.WriteFile(filepath.Join(external, "config"), []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	result := Adapter{ConfigDir: external, BoundaryRoot: boundary}.Apply(context.Background(), palette())
	got, err := os.ReadFile(filepath.Join(external, "config"))
	if err != nil || result.Status != model.AdapterConflict || string(got) != original {
		t.Fatalf("result=%#v config=%q err=%v", result, got, err)
	}
}

func TestApplyRejectsSymlinkedAncestorAndLexicalEscapeBeforeWriting(t *testing.T) {
	for _, configDir := range []string{"symlinked-ancestor", "lexical-escape"} {
		t.Run(configDir, func(t *testing.T) {
			dir := tempDir(t)
			external := tempDir(t)
			original := marker + "\ntheme = old\n"
			if err := os.WriteFile(filepath.Join(external, "config"), []byte(original), 0600); err != nil {
				t.Fatal(err)
			}
			var target string
			switch configDir {
			case "symlinked-ancestor":
				link := filepath.Join(dir, "link")
				if err := os.Symlink(external, link); err != nil {
					t.Fatal(err)
				}
				target = link
			case "lexical-escape":
				target = external + string(os.PathSeparator) + "nested" + string(os.PathSeparator) + ".."
			}
			result := Adapter{ConfigDir: target, BoundaryRoot: dir}.Apply(context.Background(), palette())
			got, err := os.ReadFile(filepath.Join(external, "config"))
			if err != nil || result.Status != model.AdapterConflict || string(got) != original {
				t.Fatalf("result=%#v config=%q err=%v", result, got, err)
			}
			if _, err := os.Stat(filepath.Join(external, "themes", "hather.conf")); !os.IsNotExist(err) {
				t.Fatalf("theme artifact created through unsafe path: %v", err)
			}
		})
	}
}

func TestApplyRejectsSymlinkedConfigDirectoryBeforeWriting(t *testing.T) {
	realDir := tempDir(t)
	if err := os.WriteFile(filepath.Join(realDir, "config"), []byte(marker+"\ntheme = old\n"), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(tempDir(t), "ghostty")
	if err := os.Symlink(realDir, link); err != nil {
		t.Fatal(err)
	}
	result := Adapter{ConfigDir: link + string(os.PathSeparator), BoundaryRoot: filepath.Dir(link)}.Apply(context.Background(), palette())
	if result.Status != model.AdapterConflict {
		t.Fatalf("result=%#v, want unsafe-path conflict", result)
	}
	if _, err := os.Stat(filepath.Join(realDir, "themes", "hather.conf")); !os.IsNotExist(err) {
		t.Fatalf("theme artifact created through symlink: %v", err)
	}
}

func TestApplyProtectsUnownedGhosttyConfigAndMissingTarget(t *testing.T) {
	for _, name := range []string{"config_theme.conf", "config_inline_palette.conf"} {
		t.Run(name, func(t *testing.T) {
			dir := tempDir(t)
			config := filepath.Join(dir, "config")
			original, err := os.ReadFile(filepath.Join("testdata", name))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(config, original, 0600); err != nil {
				t.Fatal(err)
			}
			result := Adapter{ConfigDir: dir, BoundaryRoot: dir}.Apply(context.Background(), palette())
			got, _ := os.ReadFile(config)
			if result.Status != model.AdapterConflict || string(got) != string(original) {
				t.Fatalf("result=%#v bytes changed=%t", result, string(got) != string(original))
			}
		})
	}
	boundary := tempDir(t)
	result := Adapter{ConfigDir: filepath.Join(boundary, "missing"), BoundaryRoot: boundary}.Apply(context.Background(), palette())
	if result.Status != model.AdapterUnavailable {
		t.Fatalf("missing target = %#v", result)
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

func palette() model.Palette {
	return model.Palette{Background: "#111111", Foreground: "#eeeeee", Accent: "#ff0000", Colors: [16]string{"#000000", "#111111"}}
}
