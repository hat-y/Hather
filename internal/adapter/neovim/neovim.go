package neovim

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

const Header = "\" Hather-owned colorscheme\n"

type Adapter struct {
	ConfigRoot, BoundaryRoot string
}

func DefaultConfigRoot(home string) string { return filepath.Join(home, ".config", "nvim") }
func (Adapter) ID() string                 { return "neovim" }

func (a Adapter) Apply(ctx context.Context, palette model.Palette) model.AdapterResult {
	if err := ctx.Err(); err != nil {
		return result(model.AdapterFailed, "cancelled", "Neovim colorscheme was not changed: "+err.Error())
	}
	root := filepath.Clean(a.ConfigRoot)
	dir := filepath.Join(root, "colors")
	if symlink(root) || symlink(dir) || (a.BoundaryRoot != "" && unsafeUnder(a.BoundaryRoot, root)) {
		return result(model.AdapterConflict, "config", "Neovim colors directory is a symlink and was not changed")
	}
	if _, err := os.Stat(root); errors.Is(err, os.ErrNotExist) {
		return result(model.AdapterUnavailable, "config", "Neovim config directory is unavailable; create it before selecting Neovim")
	} else if err != nil {
		return result(model.AdapterFailed, "config", "could not access Neovim config directory: "+err.Error())
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return result(model.AdapterFailed, "config", "could not create Neovim colors directory: "+err.Error())
	}
	path := filepath.Join(dir, "hather.vim")
	if existing, err := os.ReadFile(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return result(model.AdapterFailed, "config", "could not read Neovim colorscheme: "+err.Error())
	} else if err == nil && !strings.HasPrefix(string(existing), Header) {
		return result(model.AdapterConflict, "config", "existing Neovim colorscheme is not Hather-owned and was not changed")
	}
	if err := edit.Write(path, colorscheme(palette)); err != nil {
		return result(model.AdapterFailed, "config", "could not write Neovim colorscheme: "+err.Error())
	}
	return model.AdapterResult{AdapterID: a.ID(), Status: model.AdapterApplied, GeneratedArtifacts: []string{path}, FollowUp: []string{"select on next start or run :colorscheme hather"}, Diagnostics: []string{"wrote Hather colorscheme; no Neovim server was contacted"}}
}

func result(status model.AdapterStatus, class, message string) model.AdapterResult {
	return model.AdapterResult{AdapterID: "neovim", Status: status, ErrorClass: class, Diagnostics: []string{message}}
}

func symlink(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode()&os.ModeSymlink != 0
}

func unsafeUnder(boundary, target string) bool {
	boundary, target = filepath.Clean(boundary), filepath.Clean(target)
	rel, err := filepath.Rel(boundary, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return true
	}
	current := boundary
	for _, part := range strings.Split(rel, string(os.PathSeparator)) {
		if part == "." {
			continue
		}
		current = filepath.Join(current, part)
		if info, err := os.Lstat(current); err == nil && info.Mode()&os.ModeSymlink != 0 {
			return true
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			return true
		}
	}
	return false
}

func colorscheme(p model.Palette) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "%shi clear\nlet g:colors_name = \"hather\"\nhi Normal guifg=%s guibg=%s\nhi Comment guifg=%s\nhi Visual guibg=%s\n", Header, p.Foreground, p.Background, p.Muted, p.Accent)
	for i, color := range p.Colors {
		fmt.Fprintf(&b, "let g:terminal_color_%d = \"%s\"\n", i, color)
	}
	return []byte(b.String())
}
