package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hat-y/Hather/internal/app"
	"github.com/hat-y/Hather/internal/model"
	"github.com/hat-y/Hather/internal/palette"
	"github.com/hat-y/Hather/internal/source"
)

type fakeCore struct {
	searchQuery string
	previewed   model.Wallpaper
	applied     app.ApplyRequest
	result      model.OperationResult
}

func (f *fakeCore) Search(_ context.Context, query string) ([]model.Wallpaper, error) {
	f.searchQuery = query
	return []model.Wallpaper{{ID: "wall-1", Title: "Forest"}}, nil
}
func (f *fakeCore) Preview(_ context.Context, wall model.Wallpaper) model.OperationResult {
	f.previewed = wall
	return f.result
}
func (f *fakeCore) Apply(_ context.Context, request app.ApplyRequest) model.OperationResult {
	f.applied = request
	return f.result
}

func TestRunRoutesCommandsAndRendersModelJSON(t *testing.T) {
	core := &fakeCore{result: model.OperationResult{Status: model.OperationComplete, Palette: model.Palette{Accent: "#123456"}}}
	var out bytes.Buffer
	cli := CLI{Core: core, Version: "test", Out: &out}
	if code := cli.Run(context.Background(), []string{"search", "--query", "forest"}); code != 0 || core.searchQuery != "forest" {
		t.Fatalf("search code=%d query=%q", code, core.searchQuery)
	}
	out.Reset()
	if code := cli.Run(context.Background(), []string{"preview", "--image", "local.png", "--json"}); code != 0 || core.previewed.ImageURL != "local.png" {
		t.Fatalf("preview code=%d wall=%#v", code, core.previewed)
	}
	var got model.OperationResult
	if err := json.Unmarshal(out.Bytes(), &got); err != nil || got.Status != model.OperationComplete || got.Palette.Accent != "#123456" {
		t.Fatalf("JSON result=%#v err=%v", got, err)
	}
	core.result.Status = model.OperationPartialFailure
	out.Reset()
	if code := cli.Run(context.Background(), []string{"apply", "--image", "local.png", "--adapters", "ghostty"}); code != 2 || core.applied.Adapters[0] != "ghostty" || !strings.Contains(out.String(), "partial_failure") {
		t.Fatalf("apply code=%d request=%#v output=%q", code, core.applied, out.String())
	}
	out.Reset()
	if code := cli.Run(context.Background(), []string{"version"}); code != 0 || strings.TrimSpace(out.String()) != "test" {
		t.Fatalf("version code=%d output=%q", code, out.String())
	}
}

func TestHelpAndVersionAliases(t *testing.T) {
	for _, test := range []struct {
		args []string
		want string
	}{
		{[]string{"help"}, "Usage: hather"},
		{[]string{"-h"}, "Usage: hather"},
		{[]string{"--help"}, "Usage: hather"},
		{[]string{"--version"}, "v1.2.3"},
	} {
		t.Run(strings.Join(test.args, " "), func(t *testing.T) {
			var out bytes.Buffer
			program := CLI{Version: "v1.2.3", Out: &out}
			if code := program.Run(context.Background(), test.args); code != ExitComplete {
				t.Fatalf("exit code = %d", code)
			}
			if got := strings.TrimSpace(out.String()); !strings.Contains(got, test.want) {
				t.Fatalf("output = %q, want %q", got, test.want)
			}
		})
	}
}

func TestSearchWallhavenAPIKeyIsTransient(t *testing.T) {
	core := &fakeCore{}
	var out bytes.Buffer
	configured := ""
	cli := CLI{Core: core, Out: &out, ConfigureWallhavenAPIKey: func(key string) { configured = key }}
	if code := cli.Run(context.Background(), []string{"search", "--query", "forest", "--wallhaven-api-key", "flag-key", "--json"}); code != ExitComplete {
		t.Fatalf("search exit code = %d", code)
	}
	if configured != "flag-key" || core.searchQuery != "forest" {
		t.Fatalf("configured key = %q, query = %q", configured, core.searchQuery)
	}
	if strings.Contains(out.String(), "flag-key") {
		t.Fatalf("output leaked API key: %q", out.String())
	}
}

func TestWallhavenSelectionAndTextRenderingUseCompleteResultEvidence(t *testing.T) {
	core := &fakeCore{result: model.OperationResult{
		Status:      model.OperationPartialFailure,
		Wallpaper:   model.Wallpaper{Title: "Wallhaven w1", Metadata: map[string]string{"resolution": "1920x1080"}},
		Artifact:    model.Artifact{Path: "/tmp/w1.jpg", SHA256: "abc", ContentType: "image/jpeg", Size: 12},
		Diagnostics: []string{"download warning"},
		AdapterResults: []model.AdapterResult{{
			AdapterID: "ghostty", Status: model.AdapterUnavailable, ErrorClass: "unavailable",
			ChangedPaths: []string{"/tmp/config"}, GeneratedArtifacts: []string{"/tmp/theme"},
			Diagnostics: []string{"missing config"}, FollowUp: []string{"Install Ghostty"},
		}},
	}}
	var out bytes.Buffer
	configured := ""
	program := CLI{Core: core, Out: &out, ConfigureWallhavenAPIKey: func(key string) { configured = key }}
	code := program.Run(context.Background(), []string{"preview", "--wallhaven-id", "w1", "--wallhaven-image-url", "https://example.test/w1.jpg", "--wallhaven-page-url", "https://wallhaven.cc/w/w1", "--wallhaven-api-key", "flag-key"})
	if code != ExitPartial || configured != "flag-key" || core.previewed.SourceKind != "wallhaven" || core.previewed.ID != "w1" || core.previewed.ImageURL != "https://example.test/w1.jpg" || core.previewed.PageURL != "https://wallhaven.cc/w/w1" {
		t.Fatalf("code=%d key=%q wall=%#v", code, configured, core.previewed)
	}
	for _, want := range []string{"status: partial_failure", "wallpaper: Wallhaven w1", "resolution: 1920x1080", "artifact: /tmp/w1.jpg", "sha256: abc", "content type: image/jpeg", "size: 12", "ghostty: unavailable", "error class: unavailable", "changed paths: /tmp/config", "generated artifacts: /tmp/theme", "diagnostics: missing config", "follow-up: Install Ghostty", "download warning"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("missing %q in %q", want, out.String())
		}
	}
}

func TestWallhavenApplyCarriesSearchMetadataInTextAndJSON(t *testing.T) {
	wall := model.Wallpaper{SourceKind: "wallhaven", ID: "abc", Title: "Wallhaven abc", PageURL: "https://wallhaven.cc/w/abc", ImageURL: "https://example.test/a.jpg", Attribution: "Wallhaven: https://wallhaven.cc/w/abc", Metadata: map[string]string{"resolution": "1920x1080", "category": "general", "purity": "sfw"}}
	core := &attributionCore{wall: wall}
	var out bytes.Buffer
	program := CLI{Core: core, Out: &out}
	args := []string{"apply", "--wallhaven-id", wall.ID, "--wallhaven-image-url", wall.ImageURL, "--wallhaven-page-url", wall.PageURL, "--wallhaven-resolution", "1920x1080", "--wallhaven-category", "general", "--wallhaven-purity", "sfw", "--adapters", "ghostty"}
	if code := program.Run(context.Background(), args); code != 0 {
		t.Fatal(code)
	}
	for _, want := range []string{"resolution: 1920x1080", "category: general", "purity: sfw", "page: " + wall.PageURL, "attribution: " + wall.Attribution} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("apply text missing %q: %s", want, out.String())
		}
	}
	out.Reset()
	if code := program.Run(context.Background(), append(args, "--json")); code != 0 {
		t.Fatal(code)
	}
	var result model.OperationResult
	if err := json.Unmarshal(out.Bytes(), &result); err != nil || result.Wallpaper.Metadata["resolution"] != "1920x1080" || result.Wallpaper.Metadata["category"] != "general" || result.Wallpaper.Metadata["purity"] != "sfw" || result.Wallpaper.Attribution != wall.Attribution {
		t.Fatalf("apply JSON=%#v err=%v", result.Wallpaper, err)
	}
}

func TestWallhavenApplyWithoutPageDoesNotInventAttribution(t *testing.T) {
	core := &attributionCore{}
	var out bytes.Buffer
	program := CLI{Core: core, Out: &out}
	args := []string{"apply", "--wallhaven-id", "abc", "--wallhaven-image-url", "https://example.test/a.jpg", "--adapters", "ghostty"}
	if code := program.Run(context.Background(), args); code != 0 {
		t.Fatal(code)
	}
	if !strings.Contains(out.String(), "page: https://wallhaven.cc/w/abc") || !strings.Contains(out.String(), "attribution: Wallhaven: https://wallhaven.cc/w/abc") {
		t.Fatalf("missing derived attribution: %q", out.String())
	}
	out.Reset()
	if code := program.Run(context.Background(), append(args, "--json")); code != 0 {
		t.Fatal(code)
	}
	var result model.OperationResult
	if err := json.Unmarshal(out.Bytes(), &result); err != nil || result.Wallpaper.PageURL != "https://wallhaven.cc/w/abc" || result.Wallpaper.Attribution != "Wallhaven: https://wallhaven.cc/w/abc" {
		t.Fatalf("JSON attribution=%q err=%v", result.Wallpaper.Attribution, err)
	}
}

func TestWallhavenPreviewWithoutPageAndLocalWithoutAttribution(t *testing.T) {
	core := &fakeCore{result: model.OperationResult{Status: model.OperationComplete}}
	var out bytes.Buffer
	program := CLI{Core: core, Out: &out}
	program.Run(context.Background(), []string{"preview", "--wallhaven-id", "a/b", "--wallhaven-image-url", "https://example.test/a.jpg", "--json"})
	if core.previewed.PageURL != "https://wallhaven.cc/w/a%2Fb" || core.previewed.Attribution != "Wallhaven: "+core.previewed.PageURL {
		t.Fatalf("preview wallpaper=%#v", core.previewed)
	}
	out.Reset()
	program.Run(context.Background(), []string{"preview", "--image", "local.png"})
	if core.previewed.PageURL != "" || core.previewed.Attribution != "" || strings.Contains(out.String(), "attribution:") {
		t.Fatalf("local wallpaper=%#v output=%q", core.previewed, out.String())
	}
}

func TestWallhavenSearchAndApplyTextAttribution(t *testing.T) {
	wall := model.Wallpaper{SourceKind: "wallhaven", ID: "abc", Title: "Wallhaven abc", PageURL: "https://wallhaven.cc/w/abc", ImageURL: "https://example.test/a.jpg", Attribution: "Wallhaven: https://wallhaven.cc/w/abc", Metadata: map[string]string{"resolution": "1920x1080", "category": "general", "purity": "sfw"}}
	core := &attributionCore{wall: wall}
	var out bytes.Buffer
	program := CLI{Core: core, Out: &out}
	if code := program.Run(context.Background(), []string{"search", "--query", "forest"}); code != 0 {
		t.Fatal(code)
	}
	for _, want := range []string{wall.PageURL, wall.Attribution, "category: general", "purity: sfw"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("search missing %q: %s", want, out.String())
		}
	}
	out.Reset()
	if code := program.Run(context.Background(), []string{"apply", "--wallhaven-id", wall.ID, "--wallhaven-image-url", wall.ImageURL, "--wallhaven-page-url", wall.PageURL, "--adapters", "ghostty"}); code != 0 {
		t.Fatal(code)
	}
	for _, want := range []string{wall.PageURL, wall.Attribution} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("apply missing %q: %s", want, out.String())
		}
	}
}

type attributionCore struct{ wall model.Wallpaper }

func (f *attributionCore) Search(context.Context, string) ([]model.Wallpaper, error) {
	return []model.Wallpaper{f.wall}, nil
}
func (f *attributionCore) Preview(context.Context, model.Wallpaper) model.OperationResult {
	return model.OperationResult{Status: model.OperationComplete, Wallpaper: f.wall}
}
func (f *attributionCore) Apply(_ context.Context, r app.ApplyRequest) model.OperationResult {
	f.wall = r.Wallpaper
	return model.OperationResult{Status: model.OperationComplete, Wallpaper: r.Wallpaper}
}

func TestPaletteRolesInPreviewAndApplyText(t *testing.T) {
	p := model.Palette{Background: "#010203", Foreground: "#fefdfc", Muted: "#778899", Accent: "#abcdef"}
	core := &fakeCore{result: model.OperationResult{Status: model.OperationComplete, Palette: p}}
	for _, command := range []string{"preview", "apply"} {
		var out bytes.Buffer
		program := CLI{Core: core, Out: &out}
		if code := program.Run(context.Background(), []string{command, "--image", "wall.png"}); code != 0 {
			t.Fatal(code)
		}
		for _, want := range []string{"background: " + p.Background, "foreground: " + p.Foreground, "muted: " + p.Muted, "accent: " + p.Accent} {
			if !strings.Contains(out.String(), want) {
				t.Fatalf("%s missing %q: %s", command, want, out.String())
			}
		}
	}
}

func TestSearchUsesConfiguredQueryUnlessExplicit(t *testing.T) {
	core := &fakeCore{}
	program := CLI{Core: core, DefaultQuery: "forest", QuerySource: "config"}
	if code := program.Run(context.Background(), []string{"search"}); code != 0 || core.searchQuery != "forest" {
		t.Fatalf("default code=%d query=%q", code, core.searchQuery)
	}
	if code := program.Run(context.Background(), []string{"search", "--query", "ocean"}); code != 0 || core.searchQuery != "ocean" {
		t.Fatalf("explicit code=%d query=%q", code, core.searchQuery)
	}
}

func TestTextRenderingIncludesPaletteDiagnostics(t *testing.T) {
	var out bytes.Buffer
	CLI{Out: &out}.render(model.OperationResult{Palette: model.Palette{Diagnostics: []string{"low contrast", "fallback palette"}}}, false)

	for _, want := range []string{"low contrast", "fallback palette"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("missing palette diagnostic %q in %q", want, out.String())
		}
	}
}

func TestPreviewLocalImageIsOfflineAndExitCodesAreStable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wall.png")
	fixture, err := os.ReadFile("../palette/testdata/fixture.png")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, fixture, 0600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	cli := CLI{Core: app.Service{Source: source.Local{}, Engine: palette.New()}, Out: &out}
	if code := cli.Run(context.Background(), []string{"preview", "--image", path}); code != 0 || !strings.Contains(out.String(), "complete") {
		t.Fatalf("offline preview code=%d output=%q", code, out.String())
	}
	for status, want := range map[model.OperationStatus]int{
		model.OperationComplete: 0, model.OperationPartialFailure: 2, model.OperationPrerequisiteFailure: 3, model.OperationCancelled: 5,
	} {
		if got := ExitCode(model.OperationResult{Status: status}); got != want {
			t.Errorf("%s code=%d want=%d", status, got, want)
		}
	}
	if code := cli.Run(context.Background(), []string{"preview"}); code != 4 {
		t.Fatalf("usage code=%d", code)
	}
}
