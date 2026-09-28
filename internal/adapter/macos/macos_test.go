package macos

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hat-y/Hather/internal/model"
	"github.com/hat-y/Hather/internal/platform"
)

type fixture struct {
	ExitCode int      `json:"exit_code"`
	Stdout   string   `json:"stdout"`
	Stderr   string   `json:"stderr"`
	Input    string   `json:"input_path"`
	Literal  string   `json:"expected_applescript_literal"`
	Command  []string `json:"command"`
}

type fakeExecutor struct {
	run   func(string, []string) platform.CommandResult
	calls []fixture
}

func (f *fakeExecutor) Run(_ context.Context, command string, args ...string) platform.CommandResult {
	f.calls = append(f.calls, fixture{Command: append([]string{command}, args...)})
	return f.run(command, args)
}

func TestApplySetsEveryDesktopAndDarkAppearance(t *testing.T) {
	success := loadFixture(t, "exec_success.json")
	executor := &fakeExecutor{run: standardResult(success)}
	result := Adapter{Executor: executor}.Apply(context.Background(), model.Artifact{Path: "/Users/me/Pictures/wall.jpg"}, model.Palette{Background: "#101010"})

	if result.Status != model.AdapterApplied || result.AdapterID != "macos" || !strings.Contains(strings.Join(result.Diagnostics, " "), "Spaces") || !strings.Contains(strings.Join(result.Diagnostics, " "), "instantly") {
		t.Fatalf("result = %#v, want applied macos result", result)
	}
	if len(executor.calls) != 3 {
		t.Fatalf("commands = %#v, want scheduling, desktop probe, and apply", executor.calls)
	}
	script := executor.calls[2].Command[2]
	if !strings.Contains(script, "picture of every desktop") || !strings.Contains(script, "dark mode to true") {
		t.Fatalf("apply script = %q, want all desktops and dark appearance", script)
	}
}

func TestApplyClassifiesAutomationDenial(t *testing.T) {
	denial := loadFixture(t, "exec_error_1743.json")
	executor := &fakeExecutor{run: func(command string, args []string) platform.CommandResult {
		if command == "/usr/bin/osascript" && strings.Contains(args[1], "get name of every desktop") {
			return platform.CommandResult{ExitCode: denial.ExitCode, Stderr: denial.Stderr}
		}
		return platform.CommandResult{ExitCode: 1}
	}}
	result := Adapter{Executor: executor}.Apply(context.Background(), model.Artifact{Path: "/wall.jpg"}, model.Palette{Background: "#eeeeee"})

	if result.Status != model.AdapterFailed || result.ErrorClass != "permission" || !strings.Contains(strings.Join(result.Diagnostics, " "), "System Settings → Privacy & Security → Automation") {
		t.Fatalf("result = %#v, want actionable Automation permission failure", result)
	}
}

func TestApplyPreservesAppearanceScheduling(t *testing.T) {
	scheduled := loadFixture(t, "exec_error_appearance_schedule.json")
	executor := &fakeExecutor{run: func(command string, _ []string) platform.CommandResult {
		if command == "/usr/bin/defaults" {
			return platform.CommandResult{ExitCode: scheduled.ExitCode, Stdout: scheduled.Stdout}
		}
		return platform.CommandResult{}
	}}
	result := Adapter{Executor: executor}.Apply(context.Background(), model.Artifact{Path: "/wall.jpg"}, model.Palette{})

	if result.Status != model.AdapterConflict || !strings.Contains(strings.Join(result.Diagnostics, " "), "automatic appearance scheduling") || len(executor.calls) != 1 {
		t.Fatalf("result/calls = %#v/%#v, want scheduling conflict before Apple Events", result, executor.calls)
	}
}

func TestApplyFailsWhenSchedulingCheckCannotLaunch(t *testing.T) {
	executor := &fakeExecutor{run: func(command string, _ []string) platform.CommandResult {
		if command == defaults {
			return platform.CommandResult{ExitCode: -1, Err: errors.New("launch failed")}
		}
		return platform.CommandResult{}
	}}
	result := Adapter{Executor: executor}.Apply(context.Background(), model.Artifact{Path: "/wall.jpg"}, model.Palette{})

	if result.Status != model.AdapterFailed || result.ErrorClass != "platform" || len(executor.calls) != 1 {
		t.Fatalf("result/calls = %#v/%#v, want scheduling launch failure before Apple Events", result, executor.calls)
	}
}

func TestApplyFailsWhenApplyCommandCannotLaunch(t *testing.T) {
	executor := &fakeExecutor{run: func(command string, args []string) platform.CommandResult {
		if command == defaults {
			return platform.CommandResult{ExitCode: 1}
		}
		if strings.Contains(args[1], "get name of every desktop") {
			return platform.CommandResult{Stdout: "Desktop 1"}
		}
		return platform.CommandResult{ExitCode: -1, Err: errors.New("launch failed")}
	}}
	result := Adapter{Executor: executor}.Apply(context.Background(), model.Artifact{Path: "/wall.jpg"}, model.Palette{})

	if result.Status != model.AdapterFailed || result.ErrorClass != "platform" {
		t.Fatalf("result = %#v, want apply launch failure", result)
	}
}

func TestApplyRejectsMalformedDesktopProbe(t *testing.T) {
	malformed := loadFixture(t, "exec_error_malformed.json")
	executor := &fakeExecutor{run: func(command string, args []string) platform.CommandResult {
		if command == "/usr/bin/osascript" && strings.Contains(args[1], "get name of every desktop") {
			return platform.CommandResult{ExitCode: malformed.ExitCode, Stdout: malformed.Stdout}
		}
		return platform.CommandResult{ExitCode: 1}
	}}
	result := Adapter{Executor: executor}.Apply(context.Background(), model.Artifact{Path: "/wall.jpg"}, model.Palette{})

	if result.Status != model.AdapterFailed || result.ErrorClass != "platform" || len(executor.calls) != 2 {
		t.Fatalf("result/calls = %#v/%#v, want malformed preflight failure", result, executor.calls)
	}
}

func TestApplyEscapesAppleScriptPathWithoutShell(t *testing.T) {
	vector := loadFixture(t, "exec_path_escaping.json")
	executor := &fakeExecutor{run: standardResult(fixture{})}
	result := Adapter{Executor: executor}.Apply(context.Background(), model.Artifact{Path: vector.Input}, model.Palette{Background: "#ffffff"})

	if result.Status != model.AdapterApplied {
		t.Fatalf("result = %#v, want applied", result)
	}
	for _, call := range executor.calls {
		if call.Command[0] == "/bin/sh" || call.Command[0] == "sh" {
			t.Fatalf("shell command = %#v", call.Command)
		}
	}
	script := executor.calls[2].Command[2]
	if !strings.Contains(script, vector.Literal) || !strings.Contains(script, "dark mode to false") {
		t.Fatalf("apply script = %q, want escaped light-appearance update", script)
	}
}

func loadFixture(t *testing.T, name string) fixture {
	t.Helper()
	bytes, err := os.ReadFile(filepath.Join("..", "..", "platform", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	var value fixture
	if err := json.Unmarshal(bytes, &value); err != nil {
		t.Fatal(err)
	}
	return value
}

func standardResult(success fixture) func(string, []string) platform.CommandResult {
	return func(command string, args []string) platform.CommandResult {
		if command == "/usr/bin/defaults" {
			return platform.CommandResult{ExitCode: 1, Err: errors.New("exit status 1")}
		}
		if strings.Contains(args[1], "get name of every desktop") {
			return platform.CommandResult{Stdout: "Desktop 1, Desktop 2"}
		}
		return platform.CommandResult{ExitCode: success.ExitCode, Stdout: success.Stdout, Stderr: success.Stderr}
	}
}
