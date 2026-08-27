package tooling

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/destyk/repokit/internal/config"
	"github.com/destyk/repokit/internal/repository"
)

// AppliedFile is one successful tooling mapping result.
type AppliedFile struct {
	Source      string // path relative to tooling cache
	Destination string // absolute destination path
}

// Tooling manages a shared tooling repository and its mappings.
//
// Methods never write to stdout/stderr — callers (CLI) own presentation.
type Tooling struct {
	Root   string
	Config config.ToolingConfig
}

// New creates a tooling manager.
func New(root string, cfg config.ToolingConfig) Tooling {
	return Tooling{Root: root, Config: cfg}
}

// ToolingPath returns the relative cache path for shared tooling.
func (t Tooling) ToolingPath() string {
	return filepath.Join(".repokit", "tooling")
}

// RootToolingPath returns the absolute cache path for shared tooling.
func (t Tooling) RootToolingPath() string {
	return filepath.Join(t.Root, t.ToolingPath())
}

// Sync downloads or updates the shared tooling repository.
func (t Tooling) Sync(ctx context.Context) error {
	repo := repository.New(t.Root, config.Repository{
		Name:     "tooling",
		URL:      t.Config.Repository,
		Ref:      t.Config.Ref,
		RefType:  t.Config.RefType,
		Path:     t.ToolingPath(),
		Required: true,
	})
	if err := repo.Sync(ctx); err != nil {
		return fmt.Errorf("sync tooling repository: %w", err)
	}

	return nil
}

// Apply copies all configured tooling files into the project.
//
// Returns the list of applied files. Skipped optional mappings are omitted.
// Idempotent: atomic overwrite of destinations.
func (t Tooling) Apply() ([]AppliedFile, error) {
	var applied []AppliedFile
	for _, mapping := range t.Config.Files {
		files, err := t.applyFile(mapping)
		if err != nil {
			return applied, err
		}

		applied = append(applied, files...)
	}

	return applied, nil
}

// Clean removes the local tooling repository cache.
func (t Tooling) Clean() error {
	if err := os.RemoveAll(t.RootToolingPath()); err != nil {
		return fmt.Errorf("remove tooling cache: %w", err)
	}

	return nil
}

func (t Tooling) applyFile(mapping config.FileMapping) ([]AppliedFile, error) {
	matches, err := resolveFiles(t.RootToolingPath(), mapping.Source)
	if err != nil {
		return nil, fmt.Errorf("resolve tooling files %q: %w", mapping.Source, err)
	}

	if len(matches) == 0 {
		if mapping.Required {
			return nil, fmt.Errorf("required tooling files not found: %s", mapping.Source)
		}

		return nil, nil
	}

	singleFile := len(matches) == 1 && !isGlob(mapping.Source)
	var applied []AppliedFile

	for _, source := range matches {
		relative, err := filepath.Rel(t.RootToolingPath(), source)
		if err != nil {
			return applied, fmt.Errorf("resolve relative tooling path: %w", err)
		}

		var destination string
		if singleFile {
			destination = filepath.Join(t.Root, mapping.Destination)
		} else {
			destination = filepath.Join(t.Root, mapping.Destination, relative)
		}

		if err := copyFile(source, destination); err != nil {
			return applied, fmt.Errorf("copy tooling %s -> %s: %w", relative, destination, err)
		}

		applied = append(applied, AppliedFile{
			Source:      relative,
			Destination: destination,
		})
	}

	return applied, nil
}

func isGlob(pattern string) bool {
	return strings.ContainsAny(pattern, "*?[")
}

func copyFile(source, destination string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()

	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}

	temp, err := os.CreateTemp(filepath.Dir(destination), ".repokit-*")
	if err != nil {
		return err
	}

	tempPath := temp.Name()
	defer func() { _ = os.Remove(tempPath) }()

	if _, err := io.Copy(temp, input); err != nil {
		_ = temp.Close()
		return err
	}

	if err := temp.Chmod(0o644); err != nil {
		_ = temp.Close()
		return err
	}

	if err := temp.Close(); err != nil {
		return err
	}

	return os.Rename(tempPath, destination)
}

func resolveFiles(root, pattern string) ([]string, error) {
	matches, err := doublestar.FilepathGlob(
		filepath.Join(root, pattern),
		doublestar.WithFilesOnly(),
	)
	if err != nil {
		return nil, fmt.Errorf("glob %q: %w", pattern, err)
	}

	return matches, nil
}
