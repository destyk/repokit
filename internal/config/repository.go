package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

// AllowedRunners must stay in sync with runner.Default.
var AllowedRunners = []string{
	"make", "npm", "pnpm", "yarn", "bun",
	"cargo", "uv", "poetry", "just", "task", "mise", "go",
}

// Repository describes a service that belongs to a workspace.
type Repository struct {
	Name     string `yaml:"name"`
	URL      string `yaml:"url"`
	Ref      string `yaml:"ref"`
	RefType  string `yaml:"ref_type"`
	Required bool   `yaml:"required"`
	Path     string `yaml:"path"`
	Language string `yaml:"language,omitempty"`
	Runner   string `yaml:"runner"`
}

// Validate validates a single repository entry.
// Defaults: RefType=tag, Path=services/<name>.
func (r *Repository) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}

	if r.URL == "" {
		return fmt.Errorf("url is required")
	}

	if r.Ref == "" {
		return fmt.Errorf("ref is required")
	}

	if r.RefType == "" {
		r.RefType = "tag"
	}

	if r.RefType != "tag" && r.RefType != "branch" {
		return fmt.Errorf("ref_type must be tag or branch, got %q", r.RefType)
	}

	if r.Runner == "" {
		return fmt.Errorf("runner is required; allowed: %v", AllowedRunners)
	}

	if !slices.Contains(AllowedRunners, r.Runner) {
		return fmt.Errorf("runner %q is not supported; allowed: %v", r.Runner, AllowedRunners)
	}

	if r.Path == "" {
		r.Path = filepath.Join("services", r.Name)
	}

	if err := validateRelativePath(r.Path, false); err != nil {
		return fmt.Errorf("path: %w", err)
	}

	return nil
}

func validateUniqueRepositories(repositories []Repository) error {
	names := make(map[string]struct{}, len(repositories))
	paths := make(map[string]struct{}, len(repositories))

	for _, r := range repositories {
		if _, exists := names[r.Name]; exists {
			return fmt.Errorf("repositories: duplicate name %q", r.Name)
		}

		names[r.Name] = struct{}{}

		clean := filepath.Clean(r.Path)
		if _, exists := paths[clean]; exists {
			return fmt.Errorf("repositories: duplicate path %q", clean)
		}

		paths[clean] = struct{}{}
	}

	return nil
}

// PathFromRoot returns the absolute path of the repository inside root.
func (r Repository) PathFromRoot(root string) string {
	return filepath.Join(root, r.Path)
}

// IsPathAvailable reports whether the repository path exists.
func (r Repository) IsPathAvailable(root string) bool {
	_, err := os.Stat(r.PathFromRoot(root))
	return err == nil
}
