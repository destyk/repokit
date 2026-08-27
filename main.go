package main

import (
	"embed"
	"io/fs"
	"os"

	"github.com/destyk/repokit/cli"
)

//go:embed examples/standalone/.repokit.yml examples/workspace/.repokit.yml
var embeddedTemplates embed.FS

// Injected at build time via -ldflags.
var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	templates, err := fs.Sub(embeddedTemplates, "examples")
	if err != nil {
		_, _ = os.Stderr.WriteString("error: load embedded templates: " + err.Error() + "\n")
		os.Exit(1)
	}

	code := cli.Run(
		os.Args[1:],
		os.Stdin,
		os.Stdout,
		os.Stderr,
		cli.BuildInfo{
			Version:   version,
			Commit:    commit,
			BuildDate: buildDate,
		},
		templates,
	)
	os.Exit(code)
}
