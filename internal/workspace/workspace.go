package workspace

import (
	"context"
	"fmt"

	"github.com/destyk/repokit/internal/config"
	"github.com/destyk/repokit/internal/repository"
)

// Result is the outcome of syncing one repository.
type Result struct {
	Name   string
	Config config.Repository
	Err    error // nil on success
}

// Workspace manages a collection of repositories.
//
// Methods never write to stdout/stderr — callers (CLI) own presentation.
type Workspace struct {
	Root         string
	Repositories []config.Repository
}

// New creates a workspace manager.
func New(root string, repos []config.Repository) Workspace {
	return Workspace{Root: root, Repositories: repos}
}

// SyncRepositories clones or updates all configured repositories.
//
// Returns one Result per repository (including skipped optional failures).
// A non-nil error is returned only when a required repository fails.
func (w Workspace) SyncRepositories(ctx context.Context) ([]Result, error) {
	results := make([]Result, 0, len(w.Repositories))

	for _, cfg := range w.Repositories {
		repo := repository.New(w.Root, cfg)
		res := Result{Name: cfg.Name, Config: cfg}

		if err := repo.Sync(ctx); err != nil {
			res.Err = err
			results = append(results, res)

			if cfg.Required {
				return results, fmt.Errorf(
					"%s: required repository unavailable: %w",
					cfg.Name,
					err,
				)
			}

			continue
		}

		results = append(results, res)
	}

	return results, nil
}
