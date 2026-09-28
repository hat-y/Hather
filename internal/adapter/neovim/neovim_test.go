package neovim

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hat-y/Hather/internal/model"
)

func TestApplyWritesOnlyOwnedColorscheme(t *testing.T) {
	root := t.TempDir()
	plugin := filepath.Join(root, "lua", "plugins", "colors.lua")
	if err := os.MkdirAll(filepath.Dir(plugin), 0700); err != nil {
		t.Fatal(err)
	}
	const pluginConfig = "return { colorscheme = 'existing' }\n"
	if err := os.WriteFile(plugin, []byte(pluginConfig), 0600); err != nil {
		t.Fatal(err)
	}

	result := (Adapter{ConfigRoot: root}).Apply(context.Background(), palette())
	path := filepath.Join(root, "colors", "hather.vim")
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	unchanged, err := os.ReadFile(plugin)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{Header, `hi Normal guifg=#eeeeee guibg=#101010`, `hi Comment guifg=#777777`, `hi Visual guibg=#abcdef`, `let g:terminal_color_0 = "#000000"`, `let g:terminal_color_15 = "#ffffff"`} {
		if !strings.Contains(string(got), want) {
			t.Errorf("colorscheme missing %q:\n%s", want, got)
		}
	}
	if result.Status != model.AdapterApplied || len(result.GeneratedArtifacts) != 1 || result.GeneratedArtifacts[0] != path || !strings.Contains(strings.Join(result.FollowUp, " "), ":colorscheme hather") || string(unchanged) != pluginConfig {
		t.Fatalf("result=%#v plugin=%q", result, unchanged)
	}
}

func TestApplyRejectsUnownedColorschemeWithoutWriting(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "colors", "hather.vim")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	const original = "hi clear\nlet g:colors_name = 'someone-else'\n"
	if err := os.WriteFile(path, []byte(original), 0640); err != nil {
		t.Fatal(err)
	}

	result := (Adapter{ConfigRoot: root}).Apply(context.Background(), palette())
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != model.AdapterConflict || string(got) != original || !strings.Contains(strings.Join(result.Diagnostics, " "), "not changed") {
		t.Fatalf("result=%#v content=%q", result, got)
	}
}

func TestApplyUpdatesOwnedColorschemeDeterministically(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "colors", "hather.vim")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(Header+"old\n"), 0600); err != nil {
		t.Fatal(err)
	}

	result := (Adapter{ConfigRoot: root}).Apply(context.Background(), palette())
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != model.AdapterApplied || !strings.HasPrefix(string(got), Header) || strings.Count(string(got), "terminal_color_") != 16 || strings.Contains(string(got), "old") {
		t.Fatalf("result=%#v content=%q", result, got)
	}
}

func TestApplyRejectsSymlinkedAncestorWithinBoundary(t *testing.T) {
	home := t.TempDir()
	external := t.TempDir()
	if err := os.Symlink(external, filepath.Join(home, ".config")); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(home, ".config", "nvim")
	result := (Adapter{ConfigRoot: root, BoundaryRoot: home}).Apply(context.Background(), palette())
	if result.Status != model.AdapterConflict {
		t.Fatalf("result=%#v, want ancestor-symlink conflict", result)
	}
	if _, err := os.Stat(filepath.Join(external, "nvim", "colors", "hather.vim")); !os.IsNotExist(err) {
		t.Fatalf("colorscheme escaped boundary: %v", err)
	}
}

func TestApplyReturnsUnavailableWithoutCreatingMissingConfig(t *testing.T) {
	root := filepath.Join(t.TempDir(), "missing-nvim")
	result := (Adapter{ConfigRoot: root}).Apply(context.Background(), palette())
	if result.Status != model.AdapterUnavailable || result.ErrorClass != "config" || !strings.Contains(strings.Join(result.Diagnostics, " "), "unavailable") {
		t.Fatalf("result=%#v", result)
	}
	if _, err := os.Stat(filepath.Join(root, "colors", "hather.vim")); !os.IsNotExist(err) {
		t.Fatalf("missing Neovim target created artifacts: %v", err)
	}
}

func TestApplyCancelledDoesNotWrite(t *testing.T) {
	root := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := (Adapter{ConfigRoot: root}).Apply(ctx, palette())
	if result.Status != model.AdapterFailed || result.ErrorClass != "cancelled" {
		t.Fatalf("result=%#v", result)
	}
	if _, err := os.Stat(filepath.Join(root, "colors", "hather.vim")); !os.IsNotExist(err) {
		t.Fatalf("cancelled apply wrote a colorscheme: %v", err)
	}
}

func palette() model.Palette {
	return model.Palette{Background: "#101010", Foreground: "#eeeeee", Muted: "#777777", Accent: "#abcdef", Colors: [16]string{"#000000", "#ff0000", "#00ff00", "#ffff00", "#0000ff", "#ff00ff", "#00ffff", "#dddddd", "#222222", "#ff5555", "#55ff55", "#ffff55", "#5555ff", "#ff55ff", "#55ffff", "#ffffff"}}
}
