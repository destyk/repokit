// Package runner is the public adapter for the runner domain.
//
// The implementation lives in internal/runner. Programmatic callers get a
// structured Result; the CLI streams process output via Options.
package runner

import (
	"context"

	intrunner "github.com/destyk/repokit/internal/runner"
)

// Core types.
type (
	Runner   = intrunner.Runner
	Registry = intrunner.Registry
	Options  = intrunner.Options
	Result   = intrunner.Result
)

// NewRegistry builds a registry from the given runners.
func NewRegistry(runners ...Runner) *Registry {
	return intrunner.NewRegistry(runners...)
}

// Default returns the built-in registry.
func Default() *Registry {
	return intrunner.Default()
}

// Builtins returns all built-in runners.
func Builtins() []Runner {
	return intrunner.Builtins()
}

// NewExecRunner creates a process runner.
//
// scriptStyle=true prepends "run" (npm run X, bun run X, …).
func NewExecRunner(name string, scriptStyle bool) Runner {
	return intrunner.NewExecRunner(name, scriptStyle)
}

// Run looks up the named runner in the default registry and executes it.
//
// When opts.Stdout/Stderr are nil, output is captured into Result.
// When set (e.g. by the CLI), output is streamed to those writers.
func Run(
	ctx context.Context,
	runnerName, dir, target string,
	args []string,
	opts Options,
) (Result, error) {
	rn, ok := Default().Get(runnerName)
	if !ok {
		return Result{}, errUnknown(runnerName)
	}

	return rn.Run(ctx, dir, target, args, opts)
}

type unknownRunnerError struct{ name string }

func (e unknownRunnerError) Error() string {
	return "runner " + e.name + " is not registered"
}

func errUnknown(name string) error {
	return unknownRunnerError{name: name}
}
