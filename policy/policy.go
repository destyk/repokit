// Package policy is the public adapter for setup policies.
//
// Analogous to commitkit/policy: it turns high-level intent into domain
// specs that repository, tooling and workspace packages understand.
//
// CLI loads .repokit.yml via internal/config. Programmatic users compose
// policies instead of touching YAML.
package policy

import (
	"github.com/destyk/repokit/repository"
	"github.com/destyk/repokit/tooling"
)

// Standalone returns a tooling Spec suitable for a single-repository setup
// (no workspace services, no go.work).
//
// Callers still choose the tooling repository URL/ref and file mappings.
func Standalone(toolingRepo, ref string, files ...tooling.FileSpec) tooling.Spec {
	return tooling.Spec{
		Repository: toolingRepo,
		Ref:        ref,
		RefType:    "tag",
		Files:      files,
	}
}

// WorkspaceTooling is like Standalone but documents the intent for a
// multi-service workspace root.
func WorkspaceTooling(toolingRepo, ref string, files ...tooling.FileSpec) tooling.Spec {
	return Standalone(toolingRepo, ref, files...)
}

// Service builds a repository.Spec for a workspace member.
func Service(name, url, ref, runner string) repository.Spec {
	return repository.Spec{
		Name:    name,
		URL:     url,
		Ref:     ref,
		RefType: "tag",
		Runner:  runner,
		Path:    "services/" + name,
	}
}

// RequiredService is Service with Required=true.
func RequiredService(name, url, ref, runner string) repository.Spec {
	s := Service(name, url, ref, runner)
	s.Required = true
	return s
}

// Branch sets RefType to "branch" on a copy of the spec.
func Branch(s repository.Spec) repository.Spec {
	s.RefType = "branch"
	return s
}

// File is a convenience constructor for tooling.FileSpec.
func File(source, destination string, required bool) tooling.FileSpec {
	return tooling.FileSpec{
		Source:      source,
		Destination: destination,
		Required:    required,
	}
}
