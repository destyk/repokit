package status

import (
	"os"
	"path/filepath"

	"github.com/destyk/repokit/internal/config"
	"github.com/destyk/repokit/internal/repository"
)

// System is a full snapshot of the current repokit-managed state.
//
// Pure data — the CLI layer is responsible for formatting.
type System struct {
	Mode       string
	Root       string
	ConfigPath string

	Tooling ToolingStatus

	Repositories []repository.Status

	GoWork         bool
	PnpmWorkspace  bool
	GoWorkUseCount int
}

// ToolingStatus describes tooling configuration and local cache.
type ToolingStatus struct {
	Repository   string
	Ref          string
	RefType      string
	FileCount    int
	CachePath    string
	CachePresent bool
}

// Collect builds a System snapshot for the given root and config.
func Collect(root string, cfg *config.Config, configPath string) System {
	sys := System{
		Mode:       cfg.Mode,
		Root:       root,
		ConfigPath: configPath,
		Tooling: ToolingStatus{
			Repository: cfg.Tooling.Repository,
			Ref:        cfg.Tooling.Ref,
			RefType:    cfg.Tooling.RefType,
			FileCount:  len(cfg.Tooling.Files),
			CachePath:  filepath.Join(root, ".repokit", "tooling"),
		},
	}

	if _, err := os.Stat(sys.Tooling.CachePath); err == nil {
		sys.Tooling.CachePresent = true
	}

	if cfg.Mode == "workspace" {
		for _, repoCfg := range cfg.Repositories {
			repo := repository.New(root, repoCfg)
			sys.Repositories = append(sys.Repositories, repo.Status())
		}

		if _, err := os.Stat(filepath.Join(root, "go.work")); err == nil {
			sys.GoWork = true
			sys.GoWorkUseCount = countGoWorkUses(filepath.Join(root, "go.work"))
		}

		if _, err := os.Stat(filepath.Join(root, "pnpm-workspace.yaml")); err == nil {
			sys.PnpmWorkspace = true
		}
	}

	return sys
}

func countGoWorkUses(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}

	n := 0
	for _, line := range splitLines(string(data)) {
		trim := trimSpace(line)
		if len(trim) >= 3 && trim[:3] == "use" {
			n++
		}
	}

	return n
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}

	if start < len(s) {
		lines = append(lines, s[start:])
	}

	return lines
}

func trimSpace(s string) string {
	i, j := 0, len(s)
	for i < j && (s[i] == ' ' || s[i] == '\t' || s[i] == '\r') {
		i++
	}

	for j > i && (s[j-1] == ' ' || s[j-1] == '\t' || s[j-1] == '\r') {
		j--
	}

	return s[i:j]
}
