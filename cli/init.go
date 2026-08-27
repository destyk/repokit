package cli

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// runInit writes a starter .repokit.yml from embedded templates.
//
// This is a plain CLI command (like commitkit install-hook): no separate
// domain package, no public API beyond the CLI itself.
func runInit(args []string, templates fs.FS, out, _ io.Writer) error {
	mode := "workspace"
	force := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--force", "-f":
			force = true
		case "--mode":
			if i+1 >= len(args) {
				return fmt.Errorf("--mode requires a value")
			}
			mode = args[i+1]
			i++
		default:
			return fmt.Errorf("unknown init option: %s", args[i])
		}
	}

	if mode != "workspace" && mode != "standalone" {
		return fmt.Errorf("mode must be workspace or standalone, got %q", mode)
	}

	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	path := filepath.Join(root, ".repokit.yml")
	if _, err := os.Stat(path); err == nil && !force {
		return fmt.Errorf("%s already exists (use --force to overwrite)", path)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	templatePath := filepath.Join(mode, ".repokit.yml")
	content, err := fs.ReadFile(templates, templatePath)
	if err != nil {
		return fmt.Errorf("read template %s: %w", templatePath, err)
	}

	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	fmt.Fprintf(out, "[ok] wrote %s (mode=%s)\n", path, mode)
	fmt.Fprintln(out, "Next: edit .repokit.yml, then run: repokit sync")
	return nil
}
