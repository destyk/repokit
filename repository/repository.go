// Package repository is the public adapter for the repository domain.
//
// The implementation lives in internal/repository.
package repository

import (
	"context"

	intcfg "github.com/destyk/repokit/internal/config"
	intrepo "github.com/destyk/repokit/internal/repository"
)

// Spec describes a repository to synchronize.
type Spec struct {
	Name     string
	URL      string
	Ref      string
	RefType  string
	Required bool
	Path     string
	Runner   string
	Language string
}

// Core types.
type (
	Repository = intrepo.Repository
	Status     = intrepo.Status
)

// New creates a repository instance from a Spec.
func New(root string, spec Spec) Repository {
	return intrepo.New(root, toConfig(spec))
}

// Sync synchronizes the repository to the configured ref.
//
// Idempotent: clone if missing, otherwise fetch and checkout the pinned ref.
func Sync(ctx context.Context, root string, spec Spec) error {
	return intrepo.New(root, toConfig(spec)).Sync(ctx)
}

// StatusOf returns the status of a repository described by spec.
func StatusOf(root string, spec Spec) Status {
	return intrepo.New(root, toConfig(spec)).Status()
}

func toConfig(s Spec) intcfg.Repository {
	return intcfg.Repository{
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
