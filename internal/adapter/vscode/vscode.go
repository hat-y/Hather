package vscode

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/hat-y/Hather/internal/edit"
	"github.com/hat-y/Hather/internal/model"
	"github.com/tailscale/hujson"
)

type Adapter struct{ SettingsPath, BoundaryRoot string }

func DefaultSettingsPath(home string) string {
	return filepath.Join(home, "Library", "Application Support", "Code", "User", "settings.json")
}
func (Adapter) ID() string { return "vscode" }

func (a Adapter) Apply(ctx context.Context, p model.Palette) model.AdapterResult {
	if err := ctx.Err(); err != nil {
		return result(model.AdapterFailed, "cancelled", "VS Code settings were not changed: "+err.Error())
	}
	path := a.SettingsPath
	if path == "" {
		path = DefaultSettingsPath(a.BoundaryRoot)
	}
	if a.BoundaryRoot == "" || !safeUnder(a.BoundaryRoot, path) {
		return result(model.AdapterConflict, "config", "VS Code settings path is outside the trusted home and was not changed")
	}
	if _, err := os.Stat(filepath.Dir(path)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return result(model.AdapterUnavailable, "config", "VS Code settings directory is unavailable")
		}
		return result(model.AdapterFailed, "config", "could not access VS Code settings directory: "+err.Error())
	}
	if _, err := edit.Edit(path, func(in []byte) (edit.Change, error) {
		out, err := rewrite(in, p)
		return edit.Change{Bytes: out}, err
	}); err != nil {
		return result(model.AdapterConflict, "config", "VS Code settings were not changed: "+err.Error())
	}
	return model.AdapterResult{AdapterID: a.ID(), Status: model.AdapterApplied, ChangedPaths: []string{path}, FollowUp: []string{"reload the VS Code window to apply updated colors"}, Diagnostics: []string{"updated Hather-owned VS Code color customizations"}}
}

func result(status model.AdapterStatus, class, message string) model.AdapterResult {
	return model.AdapterResult{AdapterID: "vscode", Status: status, ErrorClass: class, Diagnostics: []string{message}}
}

func rewrite(in []byte, p model.Palette) ([]byte, error) {
	value, err := hujson.Parse(in)
	if err != nil {
		return nil, errors.New("invalid JSONC")
	}
	root, ok := value.Value.(*hujson.Object)
	if !ok {
		return nil, errors.New("settings root must be an object")
	}
	for _, section := range colors(p) {
		members := find(root, section.name)
		if len(members) > 1 {
			return nil, errors.New("ambiguous duplicate " + section.name)
		}
		var target *hujson.Object
		if len(members) == 0 {
			target = &hujson.Object{}
			root.Members = append(root.Members, hujson.ObjectMember{Name: stringValue(section.name), Value: hujson.Value{Value: target}})
		} else {
			var ok bool
			target, ok = members[0].Value.Value.(*hujson.Object)
			if !ok {
				return nil, errors.New(section.name + " must be an object")
			}
		}
		for _, entry := range section.entries {
			members := find(target, entry.key)
			if len(members) > 1 {
				return nil, errors.New("ambiguous duplicate " + entry.key)
			}
			if len(members) == 0 {
				target.Members = append(target.Members, hujson.ObjectMember{Name: stringValue(entry.key), Value: hujson.Value{Value: hujson.String(entry.color)}})
			} else {
				members[0].Value.Value = hujson.String(entry.color)
			}
		}
	}
	return value.Pack(), nil
}

type colorEntry struct{ key, color string }
type colorSection struct {
	name    string
	entries []colorEntry
}

func colors(p model.Palette) []colorSection {
	return []colorSection{
		{"workbench.colorCustomizations", []colorEntry{{"editor.background", p.Background}, {"editor.foreground", p.Foreground}, {"editor.lineHighlightBackground", p.Accent}}},
		{"editor.tokenColorCustomizations", []colorEntry{{"comments", p.Muted}}},
	}
}

func find(object *hujson.Object, wanted string) []*hujson.ObjectMember {
	var found []*hujson.ObjectMember
	for i := range object.Members {
		name, ok := object.Members[i].Name.Value.(hujson.Literal)
		if ok && name.String() == wanted {
			found = append(found, &object.Members[i])
		}
	}
	return found
}

func stringValue(s string) hujson.Value { return hujson.Value{Value: hujson.String(s)} }

func safeUnder(boundary, target string) bool {
	boundary, target = filepath.Clean(boundary), filepath.Clean(target)
	rel, err := filepath.Rel(boundary, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return false
	}
	current := boundary
	for _, part := range strings.Split(rel, string(os.PathSeparator)) {
		if part == "." {
			continue
		}
		current = filepath.Join(current, part)
		if info, err := os.Lstat(current); err == nil && info.Mode()&os.ModeSymlink != 0 {
			return false
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			return false
		}
	}
	return true
}
