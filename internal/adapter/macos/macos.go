package macos

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hat-y/Hather/internal/model"
	"github.com/hat-y/Hather/internal/platform"
)

const (
	osascript = "/usr/bin/osascript"
	defaults  = "/usr/bin/defaults"
)

type Adapter struct{ Executor platform.CommandExecutor }

func (Adapter) ID() string { return "macos" }

func (a Adapter) Apply(ctx context.Context, artifact model.Artifact, palette model.Palette) model.AdapterResult {
	if a.Executor == nil || artifact.Path == "" {
		return failure("platform", "macOS adapter requires an executor and wallpaper path")
	}
	schedule := a.Executor.Run(ctx, defaults, "read", "-g", "AppleInterfaceStyleSwitchesAutomatically")
	if schedule.ExitCode < 0 {
		return failure("platform", "could not inspect automatic appearance scheduling; no wallpaper or appearance change was made")
	}
	if schedule.ExitCode == 0 && strings.TrimSpace(schedule.Stdout) == "1" {
		return model.AdapterResult{AdapterID: a.ID(), Status: model.AdapterConflict, Diagnostics: []string{"automatic appearance scheduling is enabled; Hather will not disable it without an explicit approved override"}}
	}
	probe := a.Executor.Run(ctx, osascript, "-e", `tell application "System Events" to get name of every desktop`)
	if permission(probe) {
		return automationPermission(a)
	}
	if probe.Err != nil || probe.ExitCode != 0 || !validDesktops(probe.Stdout) {
		return failure("platform", "could not verify macOS desktops; no wallpaper or appearance change was made")
	}
	apply := a.Executor.Run(ctx, osascript, "-e", script(artifact.Path, dark(palette.Background)))
	if permission(apply) {
		return automationPermission(a)
	}
	if apply.Err != nil || apply.ExitCode != 0 {
		return failure("platform", "macOS wallpaper or appearance update failed")
	}
	return model.AdapterResult{AdapterID: a.ID(), Status: model.AdapterApplied, ChangedPaths: []string{artifact.Path}, Diagnostics: []string{"requested wallpaper for every desktop and explicit light/dark appearance; macOS may delay propagation across Spaces and running apps, which are not guaranteed to update instantly"}}
}

func automationPermission(a Adapter) model.AdapterResult {
	return model.AdapterResult{AdapterID: a.ID(), Status: model.AdapterFailed, ErrorClass: "permission", Diagnostics: []string{"macOS Automation permission is required: System Settings → Privacy & Security → Automation"}}
}

func failure(class, message string) model.AdapterResult {
	return model.AdapterResult{AdapterID: "macos", Status: model.AdapterFailed, ErrorClass: class, Diagnostics: []string{message}}
}

func permission(result platform.CommandResult) bool { return strings.Contains(result.Stderr, "-1743") }

func validDesktops(output string) bool {
	if strings.Contains(output, "<") {
		return false
	}
	for _, name := range strings.Split(output, ",") {
		if strings.TrimSpace(name) == "" {
			return false
		}
	}
	return output != ""
}

func script(path string, isDark bool) string {
	return fmt.Sprintf("tell application \"System Events\"\nset picture of every desktop to POSIX file \"%s\"\ntell appearance preferences\nset dark mode to %t\nend tell\nend tell", strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(path), isDark)
}

func dark(hex string) bool {
	if len(hex) != 7 || hex[0] != '#' {
		return true
	}
	value, err := strconv.ParseUint(hex[1:], 16, 32)
	return err != nil || (299*((value>>16)&255)+587*((value>>8)&255)+114*(value&255))/1000 < 128
}
