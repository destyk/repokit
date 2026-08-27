package config

import (
	"fmt"
	"strings"
)

// HooksConfig holds post-sync hooks.
type HooksConfig struct {
	PostSync []Hook `yaml:"post_sync,omitempty"`
}

// Hook is a single command executed after sync.
type Hook struct {
	Command  string   `yaml:"command"`
	Args     []string `yaml:"args,omitempty"`
	Required bool     `yaml:"required"`
}

// Validate validates hooks configuration.
func (c *HooksConfig) Validate() error {
	for i := range c.PostSync {
		if err := c.PostSync[i].Validate(); err != nil {
			return fmt.Errorf("post_sync[%d]: %w", i, err)
		}
	}

	return nil
}

// Validate validates a single hook.
func (h *Hook) Validate() error {
	if strings.TrimSpace(h.Command) == "" {
		return fmt.Errorf("command is required")
	}

	return nil
}
