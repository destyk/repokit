package cli

import (
	"fmt"
	"io"
	"io/fs"

	"github.com/destyk/repokit/internal/config"
)

// BuildInfo contains build metadata injected into the binary.
type BuildInfo struct {
	Version   string
	Commit    string
	BuildDate string
}

// Run executes the repokit CLI.
//
// The returned integer is intended to be passed to os.Exit.
func Run(
	args []string,
	in io.Reader,
	out io.Writer,
	errOut io.Writer,
	buildInfo BuildInfo,
	templates fs.FS,
) int {
	if len(args) == 0 {
		usage(out)
		return 0
	}

	switch args[0] {
	case "help", "-h", "--help":
		usage(out)
		return 0

	case "version", "-v", "--version":
		printVersion(out, buildInfo)
		return 0

	case "init":
		if err := runInit(args[1:], templates, out, errOut); err != nil {
			fmt.Fprintf(errOut, "error: %v\n", err)
			return 1
		}
		return 0
	}

	// Commands below require a project config.
	result, err := config.FindAndLoad("")
	if err != nil {
		fmt.Fprintf(errOut, "error: load config: %v\n", err)
		return 1
	}
	if !result.Found {
		fmt.Fprintf(errOut, "error: %s not found in current directory or any parent\n", config.FileName)
		return 1
	}

	cfg := &result.Config
	root := result.Root

	switch args[0] {
	case "sync":
		if err := runSync(root, cfg, out, errOut); err != nil {
			fmt.Fprintf(errOut, "error: %v\n", err)
			return 1
		}
		return 0

	case "status":
		if err := runStatus(root, cfg, result.Path, out, errOut); err != nil {
			fmt.Fprintf(errOut, "error: %v\n", err)
			return 1
		}
		return 0

	default:
		if err := runTarget(root, cfg, args[0], args[1:], in, out, errOut); err != nil {
			fmt.Fprintf(errOut, "error: %v\n", err)
			return 1
		}
		return 0
	}
}

func usage(w io.Writer) {
	fmt.Fprint(w, `repokit - Go repository and tooling synchronization

Usage:
  repokit init [--mode standalone|workspace] [--force]
  repokit sync
  repokit status
  repokit version
  repokit <service>/<target> [args...]

Commands:
  init      Write a starter .repokit.yml from templates.
  sync      Synchronize repositories and shared tooling.
  status    Show configured repository status.
  version   Print the version.

Configuration:
  Commands load .repokit.yml by walking up from the current directory.
  Use repokit init to create a starter config.

Examples:
  repokit init --mode standalone
  repokit sync
  repokit catalog/lint
  repokit users/test ./...
`)
}

func printVersion(w io.Writer, info BuildInfo) {
	fmt.Fprintf(w, "repokit %s (commit: %s built: %s)\n",
		info.Version, info.Commit, info.BuildDate)
}
