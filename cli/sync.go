package cli

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/destyk/repokit/internal/config"
	"github.com/destyk/repokit/internal/hooks"
	"github.com/destyk/repokit/internal/tooling"
	"github.com/destyk/repokit/internal/workspace"
)

// runSync performs a full synchronization.
//
// Idempotent. All console output is produced here; internal packages are silent.
func runSync(root string, cfg *config.Config, out, errOut io.Writer) error {
	ctx := context.Background()

	fmt.Fprintf(out, "repokit: %s sync...\n\n", cfg.Mode)

	tools := tooling.New(root, cfg.Tooling)

	fmt.Fprintln(out, "tooling")
	fmt.Fprintf(out, "  sync %s @ %s (%s)\n", cfg.Tooling.Repository, cfg.Tooling.Ref, cfg.Tooling.RefType)
	if err := tools.Sync(ctx); err != nil {
		return err
	}

	defer func() {
		if err := tools.Clean(); err != nil {
			fmt.Fprintf(errOut, "[warn] clean tooling cache: %v\n", err)
		}
	}()

	applied, err := tools.Apply()
	if err != nil {
		return err
	}
	for _, f := range applied {
		rel, _ := filepath.Rel(root, f.Destination)
		if rel == "" || strings.HasPrefix(rel, "..") {
			rel = f.Destination
		}
		fmt.Fprintf(out, "  [ok] %s -> %s\n", f.Source, rel)
	}
	if len(applied) == 0 {
		fmt.Fprintln(out, "  (no files applied)")
	}
	fmt.Fprintln(out)

	if cfg.Mode == "workspace" {
		fmt.Fprintln(out, "repositories")
		ws := workspace.New(root, cfg.Repositories)
		results, err := ws.SyncRepositories(ctx)
		for _, res := range results {
			if res.Err != nil {
				fmt.Fprintf(out, "  -> %-24s unavailable, skipped: %v\n", res.Name, res.Err)
				continue
			}
			fmt.Fprintf(out, "  -> %-24s ok (%s %s, runner=%s)\n",
				res.Name, res.Config.RefType, res.Config.Ref, res.Config.Runner)
		}
		if err != nil {
			return err
		}

		if err := workspace.ApplyGenerators(
			ctx,
			root,
			cfg.Repositories,
			workspace.DefaultGenerators(),
		); err != nil {
			return err
		}
		fmt.Fprintln(out, "  workspace files updated")
		fmt.Fprintln(out)
	}

	if cfg.Hooks != nil && len(cfg.Hooks.PostSync) > 0 {
		fmt.Fprintln(out, "hooks")
		h := hooks.New(root)
		results, err := h.RunPostSync(ctx, cfg.Hooks.PostSync)
		for _, res := range results {
			cmd := res.Command
			if len(res.Args) > 0 {
				cmd = cmd + " " + strings.Join(res.Args, " ")
			}
			if res.Err != nil {
				fmt.Fprintf(errOut, "  [warn] post-sync failed: %s: %v\n", cmd, res.Err)
				continue
			}
			fmt.Fprintf(out, "  [ok] %s\n", cmd)
		}
		if err != nil {
			return err
		}
		fmt.Fprintln(out)
	}

	fmt.Fprintf(out, "[ok] %s sync completed\n", cfg.Mode)
	return nil
}
