package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hat-y/Hather/internal/adapter"
	"github.com/hat-y/Hather/internal/adapter/vscode"
	"github.com/hat-y/Hather/internal/app"
	"github.com/hat-y/Hather/internal/cli"
	"github.com/hat-y/Hather/internal/config"
	"github.com/hat-y/Hather/internal/model"
	"github.com/hat-y/Hather/internal/platform"
	"github.com/hat-y/Hather/internal/source"
	"github.com/hat-y/Hather/internal/tui"
)

type fakeExecutor struct{ calls []string }

func (f *fakeExecutor) Run(_ context.Context, command string, args ...string) platform.CommandResult {
	f.calls = append(f.calls, command)
	if command == "/usr/bin/defaults" {
		return platform.CommandResult{ExitCode: 1}
	}
	if strings.Contains(strings.Join(args, " "), "get name of every desktop") {
		return platform.CommandResult{Stdout: "Desktop"}
	}
	return platform.CommandResult{}
}

func TestProgramConsumesQueryAndExposesOrdinaryProvenance(t *testing.T) {
	home := t.TempDir()
	path := config.PathsForHome(home).ConfigFile
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("query = \"forest\"\nenabled_adapters = [\"herdr\"]\nwallhaven_api_key = \"private\"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HATHER_QUERY", "")
	program, err := newProgram(home, "", &fakeExecutor{})
	if err != nil {
		t.Fatal(err)
	}
	if program.DefaultQuery != "forest" || program.QuerySource != "config" || program.AdapterSource != "config" {
		t.Fatalf("program provenance: %+v", program)
	}
	var out bytes.Buffer
	program.Out = &out
	program.Run(context.Background(), []string{"apply", "--image", "../../internal/palette/testdata/fixture.png"})
	if !strings.Contains(out.String(), "adapters source: config") || strings.Contains(out.String(), "private") {
		t.Fatalf("diagnostic=%q", out.String())
	}
}

func TestEnvironmentQueryAndTUIInitialInput(t *testing.T) {
	t.Setenv("HATHER_QUERY", "ocean")
	program, err := newProgram(t.TempDir(), "ghostty", &fakeExecutor{})
	if err != nil {
		t.Fatal(err)
	}
	if program.DefaultQuery != "ocean" || program.QuerySource != "environment" || program.AdapterSource != "environment" {
		t.Fatalf("provenance: %+v", program)
	}
	var out bytes.Buffer
	if code := run(context.Background(), []string{"tui"}, program, program.Core, func(m tea.Model) error {
		ui := m.(tui.Model)
		if ui.Input.Value() != "ocean" || ui.DefaultQuery != "ocean" {
			t.Errorf("initial query=%q fallback=%q", ui.Input.Value(), ui.DefaultQuery)
		}
		ui.Input.SetValue("")
		_, cmd := ui.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd == nil {
			t.Error("empty search did not dispatch")
		}
		return nil
	}, &out); code != cli.ExitComplete {
		t.Fatal(code)
	}
}

func TestProgramCompositionLoadsConfiguredAdapters(t *testing.T) {
	home := t.TempDir()
	paths := config.PathsForHome(home)
	if err := os.MkdirAll(filepath.Dir(paths.ConfigFile), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paths.ConfigFile, []byte("enabled_adapters = [\"herdr\"]\nwallhaven_api_key = \"config-secret\"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HERDR_CONFIG_PATH", "")
	t.Setenv("WALLHAVEN_API_KEY", "environment-secret")
	program, err := newProgram(home, "", &fakeExecutor{})
	if err != nil {
		t.Fatal(err)
	}
	program.ConfigureWallhavenAPIKey("flag-secret")
	var out bytes.Buffer
	program.Out = &out
	if code := program.Run(context.Background(), []string{"apply", "--image", "../../internal/palette/testdata/fixture.png", "--json"}); code != cli.ExitPartial {
		t.Fatalf("exit = %d", code)
	}
	var result model.OperationResult
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	state, err := os.ReadFile(paths.StateFile)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"config-secret", "environment-secret", "flag-secret"} {
		if strings.Contains(out.String(), secret) || strings.Contains(string(state), secret) {
			t.Fatalf("credential leaked: %q", secret)
		}
	}
	if len(result.RequestedAdapters) != 1 || result.RequestedAdapters[0] != "herdr" {
		t.Fatalf("result = %#v", result)
	}
}

func TestProgramCompositionResolvesAdapterPrecedence(t *testing.T) {
	for _, tc := range []struct {
		name, environment, config, flag, want string
	}{
		{"flag", "neovim", "herdr", "macos", "macos"},
		{"environment", "neovim", "herdr", "", "neovim"},
		{"config", "", "herdr", "", "herdr"},
		{"default", "", "", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			if tc.config != "" {
				path := config.PathsForHome(home).ConfigFile
				if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte("enabled_adapters = [\""+tc.config+"\"]\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			program, err := newProgram(home, tc.environment, &fakeExecutor{})
			if err != nil {
				t.Fatal(err)
			}
			var out bytes.Buffer
			program.Out = &out
			args := []string{"apply", "--image", "../../internal/palette/testdata/fixture.png", "--json"}
			if tc.flag != "" {
				args = append(args, "--adapters", tc.flag)
			}
			program.Run(context.Background(), args)
			var result model.OperationResult
			if err := json.Unmarshal(out.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if tc.want == "" && len(result.RequestedAdapters) != 0 {
				t.Fatalf("adapters = %v, want none", result.RequestedAdapters)
			}
			if tc.want != "" && (len(result.RequestedAdapters) != 1 || result.RequestedAdapters[0] != tc.want) {
				t.Fatalf("adapters = %v, want %q", result.RequestedAdapters, tc.want)
			}
		})
	}
}

func TestRunnerCompositionUsesMacOSGhosttyAndHerdrWithoutExternalEffects(t *testing.T) {
	executor := &fakeExecutor{}
	home := t.TempDir()
	herdrDir := filepath.Join(home, ".config", "herdr")
	if err := os.MkdirAll(herdrDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(herdrDir, "config.toml"), []byte("[terminal]\ndefault_shell = \"/bin/zsh\"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	result := newRunner(home, executor).Apply(context.Background(), adapter.ApplyContext{
		Artifact: model.Artifact{Path: "wallpaper.png"}, Palette: model.Palette{Background: "#000000"},
	}, []string{"herdr", "ghostty", "macos"})
	if len(result.AdapterResults) != 3 || result.AdapterResults[0].Status != model.AdapterApplied || result.AdapterResults[1].Status != model.AdapterUnavailable || result.AdapterResults[2].Status != model.AdapterApplied || result.Status != model.OperationPartialFailure || len(executor.calls) != 3 {
		t.Fatalf("result=%#v calls=%v", result, executor.calls)
	}
}

func TestRunnerCompositionReportsMissingNeovimConfigIndependently(t *testing.T) {
	home := t.TempDir()
	result := newRunner(home, &fakeExecutor{}).Apply(context.Background(), adapter.ApplyContext{Palette: model.Palette{Background: "#000", Foreground: "#fff"}}, []string{"neovim"})
	path := filepath.Join(home, ".config", "nvim", "colors", "hather.vim")
	if result.Status != model.OperationPartialFailure || len(result.AdapterResults) != 1 || result.AdapterResults[0].Status != model.AdapterUnavailable || result.AdapterResults[0].ErrorClass != "config" {
		t.Fatalf("result=%#v", result)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unexpected colorscheme path %q: %v", path, err)
	}
}

func TestRunnerCompositionUsesTrustedHomeForVSCode(t *testing.T) {
	home := t.TempDir()
	path := vscode.DefaultSettingsPath(home)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"editor.fontSize": 14}`), 0600); err != nil {
		t.Fatal(err)
	}
	result := newRunner(home, &fakeExecutor{}).Apply(context.Background(), adapter.ApplyContext{Palette: model.Palette{Background: "#101010", Foreground: "#eeeeee", Muted: "#777777", Accent: "#abcdef"}}, []string{"vscode"})
	got, _ := os.ReadFile(path)
	if result.Status != model.OperationComplete || len(result.AdapterResults) != 1 || result.AdapterResults[0].Status != model.AdapterApplied || !strings.Contains(string(got), "workbench.colorCustomizations") {
		t.Fatalf("result=%#v settings=%s", result, got)
	}
}

func TestRunDispatchesTUIThroughInjectedRunner(t *testing.T) {
	core := &app.Service{}
	var stderr bytes.Buffer
	called := false
	if got := run(context.Background(), []string{"tui"}, cli.CLI{}, core, func(model tea.Model) error {
		called = true
		ui, ok := model.(tui.Model)
		if !ok || ui.Core != core {
			t.Fatalf("model = %#v", model)
		}
		return nil
	}, &stderr); got != cli.ExitComplete || !called || stderr.Len() != 0 {
		t.Fatalf("exit=%d called=%t stderr=%q", got, called, stderr.String())
	}
}

func TestRunReportsTUIErrorsAndPreservesCLIDispatch(t *testing.T) {
	var stderr, stdout bytes.Buffer
	if got := run(context.Background(), []string{"tui"}, cli.CLI{}, &app.Service{}, func(tea.Model) error {
		return errors.New("terminal unavailable")
	}, &stderr); got != cli.ExitInternal || !strings.Contains(stderr.String(), "terminal unavailable") {
		t.Fatalf("tui exit=%d stderr=%q", got, stderr.String())
	}
	program := cli.CLI{Version: "test", Out: &stdout, Err: &stderr}
	if got := run(context.Background(), []string{"version"}, program, &app.Service{}, nil, &stderr); got != cli.ExitComplete || stdout.String() != "test\n" {
		t.Fatalf("cli exit=%d stdout=%q", got, stdout.String())
	}
	if got := run(context.Background(), nil, program, &app.Service{}, nil, &stderr); got != cli.ExitUsage || !strings.Contains(stderr.String(), "usage: hather") {
		t.Fatalf("bare exit=%d stderr=%q", got, stderr.String())
	}
}

func TestConfigureWallhavenAPIKeyPrefersExplicitInput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("apikey"); got != "flag-key" {
			t.Errorf("apikey = %q, want explicit flag key", got)
		}
		if got := r.Header.Get("Authorization"); got != "" {
			t.Errorf("authorization = %q, want no header", got)
		}
		w.Write([]byte(`{"data":[{"id":"wall-1","url":"https://wallhaven.cc/w/wall-1","path":"https://w.wallhaven.cc/wall-1.jpg"}]}`))
	}))
	defer server.Close()

	remote := &source.Wallhaven{BaseURL: server.URL, Client: server.Client()}
	configureWallhavenAPIKey(remote, "environment-key")("flag-key")
	if _, err := remote.Search(context.Background(), "forest"); err != nil {
		t.Fatal(err)
	}
}
