package workspace

import (
	"context"
	"fmt"
	"os"

	"github.com/destyk/repokit/internal/config"
)

// Service is a synchronized repository with an absolute path.
type Service struct {
	Config config.Repository
	Dir    string
}

// Generator writes/updates a root-level workspace file (go.work, pnpm-workspace.yaml, ...).
type Generator interface {
	Name() string
	Match(svc Service) bool
	Generate(ctx context.Context, root string, services []Service) error
}

// ApplyGenerators runs all generators against available services.
func ApplyGenerators(
	ctx context.Context,
	root string,
	repos []config.Repository,
	generators []Generator,
) error {
	var services []Service
	for _, cfg := range repos {
		dir := cfg.PathFromRoot(root)
		if _, err := os.Stat(dir); err != nil {
			continue
		}

		services = append(services, Service{Config: cfg, Dir: dir})
	}

	for _, gen := range generators {
		var matched []Service
		for _, svc := range services {
			if gen.Match(svc) {
				matched = append(matched, svc)
			}
		}

		if len(matched) == 0 {
			continue
		}

		if err := gen.Generate(ctx, root, matched); err != nil {
			return fmt.Errorf("%s: %w", gen.Name(), err)
		}
	}

	return nil
}

// DefaultGenerators returns built-in workspace file generators.
func DefaultGenerators() []Generator {
	return []Generator{
		NewGoGenerator(),
		NewPnpmGenerator(),
	}
}
