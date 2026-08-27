package cli

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/destyk/repokit/internal/config"
	"github.com/destyk/repokit/internal/repository"
	"github.com/destyk/repokit/internal/runner"
)

// runTarget executes a target in a configured service using its runner.
//
// Streams the child process IO to out/errOut. Internal runner stays free of
// any hard-coded terminal writes.
func runTarget(
	root string,
	cfg *config.Config,
	spec string,
	args []string,
	in io.Reader,
	out, errOut io.Writer,
) error {
	serviceName, target, ok := strings.Cut(spec, "/")
	if !ok || serviceName == "" || target == "" {
		return fmt.Errorf("invalid target %q: expected service/target", spec)
	}

	for _, repoCfg := range cfg.Repositories {
		if repoCfg.Name != serviceName {
			continue
		}

		repo := repository.New(root, repoCfg)
		if !repo.IsGitRepository() {
			return fmt.Errorf("service %q is not available", serviceName)
		}

		rn, ok := runner.Default().Get(repoCfg.Runner)
		if !ok {
			return fmt.Errorf("runner %q is not registered", repoCfg.Runner)
		}

		fmt.Fprintf(out, "repokit: %s/%s (runner=%s)\n", serviceName, target, repoCfg.Runner)

		_, err := rn.Run(
			context.Background(),
			repo.Path(),
			target,
			args,
			runner.Options{
				Stdin:  in,
				Stdout: out,
				Stderr: errOut,
			},
		)
		return err
	}

	return fmt.Errorf("service %q is not configured", serviceName)
}
