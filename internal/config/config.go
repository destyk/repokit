package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FileName is the default configuration file name.
const FileName = ".repokit.yml"

// Config is the on-disk representation of .repokit.yml.
//
// Config is internal to the module. External consumers should use the
// public domain packages (repository, tooling, workspace, policy) instead
// of loading YAML themselves.
type Config struct {
	Version      int           `yaml:"version"`
	Mode         string        `yaml:"mode"`
	Tooling      ToolingConfig `yaml:"tooling"`
	Repositories []Repository  `yaml:"repositories"`
	Hooks        *HooksConfig  `yaml:"hooks,omitempty"`
}

// LoadResult is returned by Load / FindAndLoad.
type LoadResult struct {
	Config Config
	Path   string // absolute path of the file that was loaded; empty if not found
	Root   string // directory that contains the config file; empty if not found
	Found  bool
}

// Default returns a minimal standalone Config.
func Default() Config {
	return Config{
		Version: 2,
		Mode:    "standalone",
	}
}

// Load reads and validates a .repokit.yml from an explicit path.
func Load(path string) (LoadResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return LoadResult{}, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return LoadResult{}, fmt.Errorf("parse config %s: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return LoadResult{}, fmt.Errorf("validate config %s: %w", path, err)
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}

	return LoadResult{
		Config: cfg,
		Path:   abs,
		Root:   filepath.Dir(abs),
		Found:  true,
	}, nil
}

// FindAndLoad walks up from startDir looking for .repokit.yml.
//
// If startDir is empty, the process working directory is used.
// If none is found, returns Default() with Found=false (no error).
func FindAndLoad(startDir string) (LoadResult, error) {
	if startDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return LoadResult{}, err
		}
		startDir = cwd
	}

	dir, err := filepath.Abs(startDir)
	if err != nil {
		return LoadResult{}, err
	}

	for {
		candidate := filepath.Join(dir, FileName)
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			return Load(candidate)
		}

		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return LoadResult{}, err
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return LoadResult{
				Config: Default(),
				Found:  false,
			}, nil
		}
		dir = parent
	}
}

// Validate validates the complete configuration.
//
// Every nested block is validated the same way: call Validate() on it and
// wrap the error with a field path prefix.
func (c *Config) Validate() error {
	if c.Version != 2 {
		return fmt.Errorf("unsupported version %d (want 2)", c.Version)
	}

	if c.Mode == "" {
		c.Mode = "standalone"
	}
	if c.Mode != "standalone" && c.Mode != "workspace" {
		return fmt.Errorf("mode: must be standalone or workspace, got %q", c.Mode)
	}

	if err := c.Tooling.Validate(); err != nil {
		return fmt.Errorf("tooling: %w", err)
	}

	if c.Mode == "workspace" && len(c.Repositories) == 0 {
		return errors.New("repositories: workspace mode requires at least one repository")
	}

	for i := range c.Repositories {
		if err := c.Repositories[i].Validate(); err != nil {
			return fmt.Errorf("repositories[%d]: %w", i, err)
		}
	}

	if err := validateUniqueRepositories(c.Repositories); err != nil {
		return err
	}

	if c.Hooks != nil {
		if err := c.Hooks.Validate(); err != nil {
			return fmt.Errorf("hooks: %w", err)
		}
	}

	return nil
}

// validateRelativePath prevents mappings and repository paths from escaping
// their configured root.
func validateRelativePath(value string, allowCurrent bool) error {
	clean := filepath.Clean(value)

	if filepath.IsAbs(clean) {
		return fmt.Errorf("path must be relative: %q", value)
	}

	if !allowCurrent && clean == "." {
		return errors.New("path must not point to current directory")
	}

	if clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("path must stay inside its root: %q", value)
	}

	return nil
}
