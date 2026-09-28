package herdr

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/hat-y/Hather/internal/edit"
	"github.com/hat-y/Hather/internal/model"
)

const ownedComment = "# Hather-owned theme settings"

type Adapter struct{ ConfigPath, BoundaryRoot string }

func (Adapter) ID() string { return "herdr" }

func (a Adapter) Apply(ctx context.Context, palette model.Palette) model.AdapterResult {
	if err := ctx.Err(); err != nil {
		return result(model.AdapterFailed, "cancelled", "herdr config was not changed: "+err.Error())
	}
	path := a.ConfigPath
	if override := os.Getenv("HERDR_CONFIG_PATH"); override != "" {
		path = override
	}
	if path == "" {
		return result(model.AdapterUnavailable, "config", "herdr config path is unavailable")
	}
	if err := edit.ValidatePathUnder(a.BoundaryRoot, path); err != nil {
		return result(model.AdapterConflict, "config", "herdr config was not changed: "+err.Error())
	}
	_, err := edit.Edit(path, func(in []byte) (edit.Change, error) {
		out, err := rewrite(in, palette)
		return edit.Change{Bytes: out}, err
	})
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return result(model.AdapterUnavailable, "config", "herdr config file is unavailable at "+path+"; create it or set HERDR_CONFIG_PATH")
		}
		return result(model.AdapterConflict, "config", "herdr config was not changed: "+err.Error())
	}
	return model.AdapterResult{
		AdapterID: a.ID(), Status: model.AdapterApplied, ChangedPaths: []string{path},
		FollowUp:    []string{"restart_required: restart herdr to load the updated theme"},
		Diagnostics: []string{"herdr config updated; no running process was signaled"},
	}
}

func result(status model.AdapterStatus, class, message string) model.AdapterResult {
	return model.AdapterResult{AdapterID: "herdr", Status: status, ErrorClass: class, Diagnostics: []string{message}}
}

var sectionOrder = map[string][]string{
	"theme":        {"name"},
	"theme.custom": {"sidebar_bg", "active_row_bg", "selection_bg", "panel_bg", "accent", "red", "green"},
}

func rewrite(in []byte, palette model.Palette) ([]byte, error) {
	if _, err := toml.Decode(string(in), &map[string]any{}); err != nil {
		return nil, errors.New("invalid or duplicate TOML")
	}
	values := map[string]map[string]string{
		"theme": {"name": "hather"},
		"theme.custom": {
			"sidebar_bg": palette.Background, "active_row_bg": palette.Colors[8],
			"selection_bg": palette.Accent, "panel_bg": palette.Background,
			"accent": palette.Accent, "red": palette.Colors[1], "green": palette.Colors[2],
		},
	}
	lines := strings.Split(string(in), "\n")
	out := make([]string, 0, len(lines)+10)
	seenSections := map[string]bool{}
	seenKeys := map[string]bool{}
	section, multiline := "", ""
	flush := func() {
		for _, key := range sectionOrder[section] {
			id := section + "." + key
			if !seenKeys[id] {
				out = append(out, key+" = "+strconv.Quote(values[section][key]))
				seenKeys[id] = true
			}
		}
	}
	for _, line := range lines {
		if multiline != "" {
			header := strings.TrimSpace(strings.SplitN(strings.TrimSpace(line), "#", 2)[0])
			if header == "[theme]" || header == "[theme.custom]" {
				return nil, errors.New("ambiguous theme table")
			}
			out = append(out, line)
			if strings.Count(line, multiline)%2 == 1 {
				multiline = ""
			}
			continue
		}
		trimmed := strings.TrimSpace(line)
		header := strings.TrimSpace(strings.SplitN(trimmed, "#", 2)[0])
		if strings.HasPrefix(header, "[") {
			flush()
			section = ""
			if header == "[theme]" || header == "[theme.custom]" {
				section = strings.Trim(header, "[]")
				if seenSections[section] {
					return nil, errors.New("ambiguous theme table")
				}
				seenSections[section] = true
			}
			out = append(out, line)
			continue
		}
		key, value, assignment := strings.Cut(trimmed, "=")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		delimiter := ""
		if strings.Contains(value, `"""`) {
			delimiter = `"""`
		} else if strings.Contains(value, `'''`) {
			delimiter = `'''`
		}
		if delimiter == "" && (strings.Contains(line, `"""`) || strings.Contains(line, `'''`)) {
			return nil, errors.New("unsupported multiline TOML value")
		}
		if section != "" {
			if _, owned := values[section][key]; assignment && owned {
				id := section + "." + key
				if seenKeys[id] || delimiter != "" {
					return nil, errors.New("ambiguous theme value")
				}
				indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
				line = indent + key + " = " + strconv.Quote(values[section][key])
				seenKeys[id] = true
			}
		}
		if assignment && delimiter != "" && strings.Count(value, delimiter)%2 == 1 {
			multiline = delimiter
		}
		out = append(out, line)
	}
	flush()
	for _, wanted := range []string{"theme", "theme.custom"} {
		if seenSections[wanted] {
			continue
		}
		if len(out) > 0 && out[len(out)-1] != "" {
			out = append(out, "")
		}
		out = append(out, ownedComment, "["+wanted+"]")
		section = wanted
		flush()
	}
	result := []byte(strings.Join(out, "\n"))
	if _, err := toml.Decode(string(result), &map[string]any{}); err != nil {
		return nil, errors.New("theme edit would produce ambiguous TOML")
	}
	return result, nil
}
