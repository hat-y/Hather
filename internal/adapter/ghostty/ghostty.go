package ghostty

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hat-y/Hather/internal/edit"
	"github.com/hat-y/Hather/internal/model"
)

const marker = "# Hather-managed theme selection"
const themeHeader = "# Hather-owned Ghostty theme\n"

type Adapter struct{ ConfigDir, BoundaryRoot string }

func (Adapter) ID() string { return "ghostty" }

func (a Adapter) Apply(_ context.Context, palette model.Palette) model.AdapterResult {
	dir := a.ConfigDir
	if err := edit.ValidatePathUnder(a.BoundaryRoot, dir); err != nil {
		return result(model.AdapterConflict, "Ghostty config directory is unsafe")
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return result(model.AdapterUnavailable, "Ghostty config directory is unavailable")
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return result(model.AdapterConflict, "Ghostty config directory is a symlink")
	}
	if !info.IsDir() {
		return result(model.AdapterUnavailable, "Ghostty config directory is unavailable")
	}
	config := filepath.Join(dir, "config")
	if err := edit.ValidatePathUnder(a.BoundaryRoot, config); err != nil {
		return result(model.AdapterConflict, "Ghostty config is unsafe")
	}
	configInfo, err := os.Lstat(config)
	if err != nil {
		return result(model.AdapterUnavailable, "Ghostty config is unavailable")
	}
	if configInfo.Mode()&os.ModeSymlink != 0 {
		return result(model.AdapterConflict, "Ghostty config is a symlink")
	}
	bytes, err := os.ReadFile(config)
	if err != nil {
		return result(model.AdapterUnavailable, "Ghostty config is unavailable")
	}
	if _, _, err := selection(bytes); err != nil {
		return result(model.AdapterConflict, "Ghostty has an unowned theme or inline palette; add the Hather marker before selecting hather")
	}
	theme := filepath.Join(dir, "themes", "hather.conf")
	if err := edit.ValidatePathUnder(a.BoundaryRoot, theme); err != nil {
		return result(model.AdapterConflict, "Ghostty theme path is unsafe")
	}
	previousTheme, previousErr := os.ReadFile(theme)
	if previousErr != nil && !os.IsNotExist(previousErr) {
		return result(model.AdapterConflict, "existing Ghostty theme is unsafe to replace")
	}
	if previousErr == nil && !strings.HasPrefix(string(previousTheme), themeHeader) {
		return result(model.AdapterConflict, "existing Ghostty theme is not Hather-owned; leave themes/hather.conf unchanged")
	}
	if err := os.MkdirAll(filepath.Dir(theme), 0700); err != nil || edit.Write(theme, themeFile(palette)) != nil {
		return result(model.AdapterFailed, "could not write Hather Ghostty theme")
	}
	changed, err := edit.Edit(config, func(in []byte) (edit.Change, error) {
		out, prior, err := selection(in)
		return edit.Change{Bytes: out, Prior: prior}, err
	})
	if err != nil {
		if previousErr == nil {
			_ = edit.Write(theme, previousTheme)
		} else {
			_ = os.Remove(theme)
		}
		return result(model.AdapterConflict, "Ghostty config changed and is no longer safe to edit")
	}
	return model.AdapterResult{AdapterID: a.ID(), Status: model.AdapterApplied, ChangedPaths: []string{config}, GeneratedArtifacts: []string{theme}, Diagnostics: []string{"Ghostty reload behavior depends on the installed version", changed.Prior}}
}

func result(status model.AdapterStatus, message string) model.AdapterResult {
	return model.AdapterResult{AdapterID: "ghostty", Status: status, Diagnostics: []string{message}}
}

func selection(in []byte) ([]byte, string, error) {
	lines := strings.Split(string(in), "\n")
	marked := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == marker {
			if marked >= 0 || i+1 >= len(lines) {
				return nil, "", errors.New("ambiguous marker")
			}
			key, _, ok := strings.Cut(strings.TrimSpace(lines[i+1]), "=")
			if !ok || strings.TrimSpace(key) != "theme" {
				return nil, "", errors.New("invalid marked theme")
			}
			marked = i + 1
			continue
		}
		key, _, ok := strings.Cut(trimmed, "=")
		key = strings.TrimSpace(key)
		if !ok || strings.HasPrefix(key, "#") {
			continue
		}
		if key == "theme" && i != marked {
			return nil, "", errors.New("unowned theme")
		}
		if key == "palette" || key == "background" || key == "foreground" {
			return nil, "", errors.New("inline palette")
		}
	}
	if marked >= 0 {
		prior := lines[marked]
		lines[marked] = "theme = hather"
		return []byte(strings.Join(lines, "\n")), prior, nil
	}
	return []byte(strings.TrimSuffix(string(in), "\n") + "\n" + marker + "\ntheme = hather\n"), "", nil
}

func themeFile(p model.Palette) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, themeHeader+"background = %s\nforeground = %s\ncursor-color = %s\n", p.Background, p.Foreground, p.Accent)
	for i, color := range p.Colors {
		fmt.Fprintf(&b, "palette = %d=%s\n", i, color)
	}
	return []byte(b.String())
}
