package platform

import (
	"context"
	"testing"
)

func TestExecutorPassesArgumentsWithoutShell(t *testing.T) {
	result := Executor{}.Run(context.Background(), "/usr/bin/printf", "%s", "$(not-a-shell)")
	if result.Err != nil || result.ExitCode != 0 || result.Stdout != "$(not-a-shell)" {
		t.Fatalf("Run() = %#v, want literal successful output", result)
	}
}

func TestExecutorMarksLaunchErrorsAsFailed(t *testing.T) {
	result := Executor{}.Run(context.Background(), "")
	if result.Err == nil || result.ExitCode == 0 {
		t.Fatalf("Run() = %#v, want a non-zero exit code for launch error", result)
	}
}
