package command

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Option configures an external command.
type Option func(*exec.Cmd, *bytes.Buffer)

// Quiet suppresses stdout; stderr is captured and attached to the error on failure.
func Quiet() Option {
	return func(cmd *exec.Cmd, stderr *bytes.Buffer) {
		cmd.Stdout = io.Discard
		cmd.Stderr = stderr
		cmd.Stdin = nil
	}
}

// WithIO attaches the given streams to the command.
//
// Pass nil to keep a stream disconnected (or already set by another option).
func WithIO(in io.Reader, out, errOut io.Writer) Option {
	return func(cmd *exec.Cmd, _ *bytes.Buffer) {
		if in != nil {
			cmd.Stdin = in
		}

		if out != nil {
			cmd.Stdout = out
		}

		if errOut != nil {
			cmd.Stderr = errOut
		}
	}
}

// InheritIO attaches the process standard streams.
// Intended for interactive service runners (make, npm, …) invoked from the CLI.
func InheritIO() Option {
	return WithIO(os.Stdin, os.Stdout, os.Stderr)
}

// Run executes an external command.
//
// By default stdout/stderr/stdin are discarded so library callers never
// accidentally write to the process terminal. Use Quiet, WithIO or InheritIO
// to change stream behaviour.
func Run(ctx context.Context, dir, name string, args []string, opts ...Option) error {
	cmd := exec.CommandContext(ctx, name, args...)
	if dir != "" {
		cmd.Dir = dir
	}

	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	cmd.Stdin = nil

	var stderr bytes.Buffer
	for _, opt := range opts {
		opt(cmd, &stderr)
	}

	if err := cmd.Run(); err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return fmt.Errorf("%s %v: %w\n%s", name, args, err, msg)
		}

		return fmt.Errorf("%s %v: %w", name, args, err)
	}

	return nil
}
