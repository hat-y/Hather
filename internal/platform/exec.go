package platform

import (
	"bytes"
	"context"
	"os/exec"
)

type CommandResult struct {
	Stdout, Stderr string
	ExitCode       int
	Err            error
}

type CommandExecutor interface {
	Run(context.Context, string, ...string) CommandResult
}

type Executor struct{}

func (Executor) Run(ctx context.Context, command string, args ...string) CommandResult {
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	result := CommandResult{Stdout: stdout.String(), Stderr: stderr.String(), Err: err}
	if err != nil {
		result.ExitCode = -1
	}
	if exit, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exit.ExitCode()
	}
	return result
}
