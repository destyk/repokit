// Package tooling is the public adapter for the tooling domain.
package tooling

import (
	"context"

	intcfg "github.com/destyk/repokit/internal/config"
	inttooling "github.com/destyk/repokit/internal/tooling"
)

// FileSpec describes a single tooling file mapping.
type FileSpec struct {
	Source      string
	Destination string
	Required    bool
}

// Spec describes shared tooling to pull and apply.
type Spec struct {
	Repository string
	Ref        string
	RefType    string
	Files      []FileSpec
}

// AppliedFile is one successful tooling mapping result.
type AppliedFile = inttooling.AppliedFile

// Tooling manages a shared tooling repository and its mappings.
type Tooling = inttooling.Tooling

// New creates a tooling manager.
func New(root string, spec Spec) Tooling {
	return inttooling.New(root, toConfig(spec))
}

// Sync downloads or updates the shared tooling repository.
func Sync(ctx context.Context, root string, spec Spec) error {
	return inttooling.New(root, toConfig(spec)).Sync(ctx)
}

// Apply copies all configured tooling files into the project.
func Apply(root string, spec Spec) ([]AppliedFile, error) {
	return inttooling.New(root, toConfig(spec)).Apply()
}

// Clean removes the local tooling repository cache.
func Clean(root string, spec Spec) error {
	return inttooling.New(root, toConfig(spec)).Clean()
}

func toConfig(s Spec) intcfg.ToolingConfig {
	files := make([]intcfg.FileMapping, len(s.Files))

	for i, f := range s.Files {
		files[i] = intcfg.FileMapping{
			Source:      f.Source,
			Destination: f.Destination,
			Required:    f.Required,
		}
	}

	return intcfg.ToolingConfig{
		Repository: s.Repository,
		Ref:        s.Ref,
		RefType:    s.RefType,
		Files:      files,
	}
}
