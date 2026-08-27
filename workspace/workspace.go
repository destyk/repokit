// Package workspace is the public adapter for the workspace domain.
package workspace

import (
	"context"

	intcfg "github.com/destyk/repokit/internal/config"
	intws "github.com/destyk/repokit/internal/workspace"
	"github.com/destyk/repokit/repository"
)

// Core types.
type (
	Workspace = intws.Workspace
	Service   = intws.Service
	Generator = intws.Generator
	Result    = intws.Result
)

// New creates a workspace manager from repository specs.
func New(root string, specs []repository.Spec) Workspace {
	return intws.New(root, toRepos(specs))
}

// SyncRepositories clones or updates all configured repositories.
func SyncRepositories(
	ctx context.Context,
	root string,
	specs []repository.Spec,
) ([]Result, error) {
	return intws.New(
		root,
		toRepos(specs),
	).SyncRepositories(ctx)
}

// ApplyGenerators runs all generators against available services.
func ApplyGenerators(
	ctx context.Context,
	root string,
	specs []repository.Spec,
	generators []Generator,
) error {
	return intws.ApplyGenerators(ctx, root, toRepos(specs), generators)
}

// DefaultGenerators returns the built-in workspace generators.
func DefaultGenerators() []Generator {
	return intws.DefaultGenerators()
}

func toRepos(specs []repository.Spec) []intcfg.Repository {
	out := make([]intcfg.Repository, len(specs))
	for i, s := range specs {
		out[i] = intcfg.Repository{
			Name:     s.Name,
			URL:      s.URL,
			Ref:      s.Ref,
			RefType:  s.RefType,
			Required: s.Required,
			Path:     s.Path,
			Runner:   s.Runner,
			Language: s.Language,
		}
	}

	return out
}
