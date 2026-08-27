package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/destyk/repokit/internal/command"
	"golang.org/x/mod/semver"
)

// GoGenerator creates go.work from services that have go.mod.
// Version = max of all go directives.
type goGenerator struct{}

// NewGoGenerator creates a generator.
func NewGoGenerator() Generator {
	return goGenerator{}
}

func (goGenerator) Name() string {
	return "go.work"
}

func (goGenerator) Match(svc Service) bool {
	_, err := os.Stat(filepath.Join(svc.Dir, "go.mod"))
	return err == nil
}

func (goGenerator) Generate(ctx context.Context, root string, services []Service) error {
	type mod struct {
		rel     string
		version string
	}

	modules := make([]mod, 0, len(services))
	for _, svc := range services {
		version, err := readGoVersion(filepath.Join(svc.Dir, "go.mod"))
		if err != nil {
			return fmt.Errorf("%s: %w", svc.Config.Name, err)
		}

		rel, err := filepath.Rel(root, svc.Dir)
		if err != nil {
			return err
		}

		modules = append(modules, mod{
			rel:     filepath.ToSlash(rel),
			version: version,
		})
	}

	goVersion := modules[0].version
	for _, m := range modules[1:] {
		if semver.Compare("v"+m.version, "v"+goVersion) > 0 {
			goVersion = m.version
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "go %s\n\nuse (\n", goVersion)
	for _, m := range modules {
		fmt.Fprintf(&b, "\t./%s\n", m.rel)
	}
	b.WriteString(")\n")

	path := filepath.Join(root, "go.work")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return fmt.Errorf("write go.work: %w", err)
	}

	return command.Run(ctx, root, "go", []string{"work", "sync"}, command.Quiet())
}

func readGoVersion(modPath string) (string, error) {
	data, err := os.ReadFile(modPath)
	if err != nil {
		return "", err
	}

	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "go ") {
			continue
		}

		v := strings.TrimSpace(strings.TrimPrefix(line, "go "))
		if i := strings.Index(v, "//"); i >= 0 {
			v = strings.TrimSpace(v[:i])
		}

		if v != "" {
			return v, nil
		}
	}

	return "", fmt.Errorf("go directive not found in %s", modPath)
}
