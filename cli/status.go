package cli

import (
	"fmt"
	"io"

	"github.com/destyk/repokit/internal/config"
	"github.com/destyk/repokit/internal/repository"
	"github.com/destyk/repokit/internal/status"
)

// runStatus prints a full system status report.
//
// Data from internal/status.Collect; formatting only here.
func runStatus(root string, cfg *config.Config, configPath string, out, _ io.Writer) error {
	sys := status.Collect(root, cfg, configPath)
	writeSystemStatus(out, sys)
	return nil
}

func writeSystemStatus(w io.Writer, sys status.System) {
	fmt.Fprintf(w, "repokit status\n")
	fmt.Fprintf(w, "─────────────\n")
	fmt.Fprintf(w, "mode:        %s\n", sys.Mode)
	fmt.Fprintf(w, "root:        %s\n", sys.Root)
	if sys.ConfigPath != "" {
		fmt.Fprintf(w, "config:      %s\n", sys.ConfigPath)
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "tooling")
	fmt.Fprintf(w, "  repository: %s\n", sys.Tooling.Repository)
	fmt.Fprintf(w, "  ref:        %s (%s)\n", sys.Tooling.Ref, sys.Tooling.RefType)
	fmt.Fprintf(w, "  mappings:   %d\n", sys.Tooling.FileCount)
	cacheState := "absent"
	if sys.Tooling.CachePresent {
		cacheState = "present"
	}
	fmt.Fprintf(w, "  cache:      %s (%s)\n", sys.Tooling.CachePath, cacheState)
	fmt.Fprintln(w)

	if sys.Mode != "workspace" {
		fmt.Fprintln(w, "repositories: (standalone — none configured)")
		return
	}

	fmt.Fprintln(w, "workspace files")
	fmt.Fprintf(w, "  go.work:              %s\n", yn(sys.GoWork))
	if sys.GoWork {
		fmt.Fprintf(w, "  go.work use entries:  %d\n", sys.GoWorkUseCount)
	}
	fmt.Fprintf(w, "  pnpm-workspace.yaml:  %s\n", yn(sys.PnpmWorkspace))
	fmt.Fprintln(w)

	if len(sys.Repositories) == 0 {
		fmt.Fprintln(w, "repositories: none configured")
		return
	}

	fmt.Fprintf(w, "repositories (%d)\n", len(sys.Repositories))
	for _, r := range sys.Repositories {
		writeRepoStatus(w, r)
	}
}

func writeRepoStatus(w io.Writer, r repository.Status) {
	req := ""
	if r.Required {
		req = ", required"
	}

	fmt.Fprintf(w, "  %s\n", r.Name)
	fmt.Fprintf(w, "    path:        %s\n", r.Path)
	fmt.Fprintf(w, "    remote:      %s\n", r.URL)
	fmt.Fprintf(w, "    pin:         %s %s%s\n", r.RefType, r.Ref, req)
	fmt.Fprintf(w, "    runner:      %s\n", r.Runner)
	if r.Language != "" {
		fmt.Fprintf(w, "    language:    %s\n", r.Language)
	}

	state := "missing"
	switch {
	case r.Available:
		state = "synced"
	case r.Exists && !r.IsGit:
		state = "exists (not a git repo)"
	case r.Exists:
		state = "present"
	}
	fmt.Fprintf(w, "    state:       %s\n", state)

	if r.Available {
		headLine := r.Head
		if r.HeadSHA != "" {
			headLine = fmt.Sprintf("%s (%s)", r.Head, r.HeadSHA)
		}
		fmt.Fprintf(w, "    head:        %s\n", headLine)
		fmt.Fprintf(w, "    matches pin: %s\n", yn(r.MatchesRef))
	}
	if r.Error != "" {
		fmt.Fprintf(w, "    note:        %s\n", r.Error)
	}
}

func yn(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}
