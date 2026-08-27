package repository

import (
	"os/exec"
	"strings"
)

// Status describes the current repository state.
//
// Pure data — formatting belongs to the CLI layer.
type Status struct {
	Name       string
	Path       string
	URL        string
	Ref        string
	RefType    string
	Runner     string
	Language   string
	Required   bool
	Available  bool // directory exists and is a git repo
	Exists     bool
	IsGit      bool
	Head       string // current branch, "detached", or empty
	HeadSHA    string
	MatchesRef bool
	Error      string
}

// Status returns repository status information.
func (r Repository) Status() Status {
	st := Status{
		Name:     r.Config.Name,
		Path:     r.Path(),
		URL:      r.Config.URL,
		Ref:      r.Config.Ref,
		RefType:  r.Config.RefType,
		Runner:   r.Config.Runner,
		Language: r.Config.Language,
		Required: r.Config.Required,
		Exists:   r.Exists(),
		IsGit:    r.IsGitRepository(),
	}
	st.Available = st.IsGit

	if !st.Available {
		return st
	}

	if out, err := exec.Command("git", "-C", r.Path(), "branch", "--show-current").Output(); err == nil {
		st.Head = strings.TrimSpace(string(out))
		if st.Head == "" {
			st.Head = "detached"
		}
	}

	if out, err := exec.Command("git", "-C", r.Path(), "rev-parse", "--short", "HEAD").Output(); err == nil {
		st.HeadSHA = strings.TrimSpace(string(out))
	}

	st.MatchesRef = headMatchesRef(r.Path(), r.Config.RefType, r.Config.Ref)

	return st
}

func headMatchesRef(dir, refType, ref string) bool {
	switch refType {
	case "branch":
		out, err := exec.Command("git", "-C", dir, "branch", "--show-current").Output()
		if err != nil {
			return false
		}
		return strings.TrimSpace(string(out)) == ref
	case "tag":
		tagSHA, err := exec.Command("git", "-C", dir, "rev-list", "-n", "1", ref).Output()
		if err != nil {
			return false
		}
		headSHA, err := exec.Command("git", "-C", dir, "rev-parse", "HEAD").Output()
		if err != nil {
			return false
		}
		return strings.TrimSpace(string(tagSHA)) == strings.TrimSpace(string(headSHA))
	default:
		return false
	}
}
