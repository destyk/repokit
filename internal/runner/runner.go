package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Runner executes a target inside a service directory.
// Name() must match config.Repository.Runner.
type Runner interface {
	Name() string
	// Run executes target with args in dir.
	//
	// IO behaviour is controlled by opts:
	//   - if Stdout/Stderr are non-nil, output is streamed there
	//   - if nil, output is captured into Result.Stdout / Result.Stderr
	Run(ctx context.Context, dir, target string, args []string, opts Options) (Result, error)
}

// Options controls stream handling for a single Run.
type Options struct {
	Stdin  io.Reader
	Stdout io.Writer // nil → capture into Result.Stdout
	Stderr io.Writer // nil → capture into Result.Stderr
}

// Result is the outcome of a runner invocation.
//
// When streams were captured (Options.Stdout/Stderr nil), Stdout and Stderr
// hold the collected text. When streams were provided by the caller, those
// fields are empty — the caller already received the bytes.
type Result struct {
	Runner   string   // e.g. "make", "npm"
	Args     []string // full argv after the binary name
	Dir      string
	ExitCode int
	Stdout   string
	Stderr   string
}

// Registry maps runner name → implementation.
type Registry struct {
	byName map[string]Runner
}

// NewRegistry builds a registry from the given runners.
func NewRegistry(runners ...Runner) *Registry {
	r := &Registry{byName: make(map[string]Runner, len(runners))}
	for _, rn := range runners {
		r.byName[rn.Name()] = rn
	}

	return r
}

// Get returns the runner registered under name.
func (r *Registry) Get(name string) (Runner, bool) {
	rn, ok := r.byName[name]
	return rn, ok
}

// Names returns registered runner names (order is not stable).
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.byName))
	for name := range r.byName {
		names = append(names, name)
	}

	return names
}

// Default is the built-in registry.
func Default() *Registry {
	return NewRegistry(Builtins()...)
}

// execRunner is a generic process-based runner.
type execRunner struct {
	name      string
	buildArgs func(target string, args []string) []string
}

// NewExecRunner creates a process runner.
//
// scriptStyle=true prepends "run" (npm run X, bun run X, …).
func NewExecRunner(name string, scriptStyle bool) Runner {
	build := directArgs
	if scriptStyle {
		build = scriptArgs
	}

	return execRunner{name: name, buildArgs: build}
}

func (e execRunner) Name() string {
	return e.name
}

func (e execRunner) Run(
	ctx context.Context,
	dir, target string,
	args []string,
	opts Options,
) (Result, error) {
	argv := e.buildArgs(target, args)
	res := Result{
		Runner: e.name,
		Args:   argv,
		Dir:    dir,
	}

	cmd := exec.CommandContext(ctx, e.name, argv...)
	if dir != "" {
		cmd.Dir = dir
	}

	var stdoutBuf, stderrBuf bytes.Buffer

	if opts.Stdin != nil {
		cmd.Stdin = opts.Stdin
	}

	if opts.Stdout != nil {
		cmd.Stdout = opts.Stdout
	} else {
		cmd.Stdout = &stdoutBuf
	}

	if opts.Stderr != nil {
		cmd.Stderr = opts.Stderr
	} else {
		cmd.Stderr = &stderrBuf
	}

	err := cmd.Run()

	// Populate captured buffers only when we owned the streams.
	if opts.Stdout == nil {
		res.Stdout = stdoutBuf.String()
	}
	if opts.Stderr == nil {
		res.Stderr = stderrBuf.String()
	}

	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			res.ExitCode = ee.ExitCode()
		} else {
			res.ExitCode = -1
		}

		msg := strings.TrimSpace(res.Stderr)
		if msg == "" && opts.Stderr != nil {
			// stderr was streamed; nothing to attach
			return res, fmt.Errorf("%s %v: %w", e.name, argv, err)
		}

		if msg != "" {
			return res, fmt.Errorf("%s %v: %w\n%s", e.name, argv, err, msg)
		}

		return res, fmt.Errorf("%s %v: %w", e.name, argv, err)
	}

	res.ExitCode = 0
	return res, nil
}

func scriptArgs(target string, args []string) []string {
	return append([]string{"run", target}, args...)
}

func directArgs(target string, args []string) []string {
	return append([]string{target}, args...)
}
