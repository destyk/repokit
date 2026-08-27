package hooks

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/destyk/repokit/internal/config"
)

// Result is the outcome of one hook execution.
type Result struct {
	Command string
	Args    []string
	Err     error // nil on success
}

// Runner executes repository hooks.
//
// Methods never write to stdout/stderr of the parent process for logging.
// Hook subprocess streams are discarded unless the caller needs otherwise
// (hooks run with discarded IO so the CLI controls all console output).
type Runner struct {
	Root string
}

// New creates a hook runner.
func New(root string) Runner {
	return Runner{Root: root}
}

// RunPostSync executes all configured post-sync hooks.
//
// Returns results for every hook. A non-nil error is returned only when a
// required hook fails.
func (r Runner) RunPostSync(ctx context.Context, hooks []config.Hook) ([]Result, error) {
	results := make([]Result, 0, len(hooks))

	for _, hook := range hooks {
		res := Result{Command: hook.Command, Args: hook.Args}
		if err := r.run(ctx, hook); err != nil {
			res.Err = err
			results = append(results, res)
			if hook.Required {
				return results, err
			}
			continue
		}
		results = append(results, res)
	}

	return results, nil
}

func (r Runner) run(ctx context.Context, hook config.Hook) error {
	if hook.Command == "" {
		return fmt.Errorf("post-sync hook command is empty")
	}

	command := r.resolveCommand(hook.Command)
	cmd := exec.CommandContext(ctx, command, hook.Args...)
	cmd.Dir = r.Root
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	cmd.Stdin = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run post-sync hook %q: %w", formatCommand(hook), err)
	}

	return nil
}

func (r Runner) resolveCommand(command string) string {
	local := filepath.Join(r.Root, filepath.FromSlash(command))
	if info, err := os.Stat(local); err == nil && !info.IsDir() {
		return local
	}

	return command
}

func formatCommand(hook config.Hook) string {
	if len(hook.Args) == 0 {
		return hook.Command
	}

	return hook.Command + " " + strings.Join(hook.Args, " ")
}
