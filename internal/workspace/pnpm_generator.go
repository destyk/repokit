package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PnpmGenerator writes pnpm-workspace.yaml for runner: pnpm services.
type pnpmGenerator struct{}

// NewPnpmGenerator creates a generator.
func NewPnpmGenerator() Generator {
	return pnpmGenerator{}
}

func (pnpmGenerator) Name() string {
	return "pnpm-workspace.yaml"
}

func (pnpmGenerator) Match(svc Service) bool {
	return svc.Config.Runner == "pnpm"
}

func (pnpmGenerator) Generate(_ context.Context, root string, services []Service) error {
	var b strings.Builder
	b.WriteString("packages:\n")
	for _, svc := range services {
		rel, err := filepath.Rel(root, svc.Dir)
		if err != nil {
			return err
		}

		fmt.Fprintf(&b, "  - '%s'\n", filepath.ToSlash(rel))
	}

	path := filepath.Join(root, "pnpm-workspace.yaml")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return fmt.Errorf("write pnpm-workspace.yaml: %w", err)
	}

	return nil
}
