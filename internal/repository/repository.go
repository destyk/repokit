package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/destyk/repokit/internal/command"
	"github.com/destyk/repokit/internal/config"
)

// Repository represents a synchronized repository.
type Repository struct {
	Config config.Repository
	Root   string
}

// New creates a repository instance from configuration.
func New(root string, cfg config.Repository) Repository {
	return Repository{
		Config: cfg,
		Root:   root,
	}
}

// Path returns the repository absolute path.
func (r Repository) Path() string {
	return filepath.Join(r.Root, r.Config.Path)
}

// Exists reports whether the repository directory exists.
func (r Repository) Exists() bool {
	_, err := os.Stat(r.Path())
	return err == nil
}

// IsGitRepository reports whether the destination is a Git repository.
func (r Repository) IsGitRepository() bool {
	_, err := os.Stat(filepath.Join(r.Path(), ".git"))
	return err == nil
}

// ValidateDestination makes sure an existing path is usable.
func (r Repository) ValidateDestination() error {
	if !r.Exists() {
		return nil
	}

	if !r.IsGitRepository() {
		return errors.New("path exists but is not a Git repository")
	}

	return nil
}

// Sync synchronizes the repository to the configured ref.
//
// Idempotent: clones if missing, otherwise fetches and checks out the pin.
// Does not write to stdout/stderr.
func (r Repository) Sync(ctx context.Context) error {
	if !r.Exists() {
		return r.clone(ctx)
	}

	if err := r.ValidateDestination(); err != nil {
		return err
	}

	return r.syncRef(ctx)
}

func (r Repository) clone(ctx context.Context) error {
	if err := os.MkdirAll(filepath.Dir(r.Path()), 0o755); err != nil {
		return fmt.Errorf("create repository parent: %w", err)
	}

	if err := command.GitClone(ctx, r.Config.URL, r.Config.Ref, r.Path()); err != nil {
		return fmt.Errorf("clone %s: %w", r.Config.Name, err)
	}

	return nil
}

func (r Repository) syncRef(ctx context.Context) error {
	switch r.Config.RefType {
	case "branch":
		if err := command.GitFetchBranch(ctx, r.Path(), r.Config.Ref); err != nil {
			return err
		}
		return command.GitCheckoutBranch(ctx, r.Path(), r.Config.Ref)
	case "tag":
		if err := command.GitFetchTags(ctx, r.Path()); err != nil {
			return err
		}
		return command.GitCheckoutTag(ctx, r.Path(), r.Config.Ref)
	default:
		return fmt.Errorf("unsupported ref type: %s", r.Config.RefType)
	}
}
