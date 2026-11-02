package app

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/0xengr/taskgo/internal/domain"
)

// ProjectTarget says how capture should choose a Project.
type ProjectTarget int

const (
	// TargetAuto infers the Project from the current directory.
	TargetAuto ProjectTarget = iota
	// TargetInbox bypasses detection.
	TargetInbox
	// TargetProject uses an explicitly named Project.
	TargetProject
)

// DeepestProjectForPath returns the Project with the deepest repository path
// containing path, regardless of archive state. It returns nil when no
// registered path contains the directory.
func DeepestProjectForPath(projects []domain.Project, path string) *domain.Project {
	var best *domain.Project
	bestLength := -1

	for i := range projects {
		project := &projects[i]
		if project.RepoPath == nil {
			continue
		}
		if !pathContains(*project.RepoPath, path) {
			continue
		}
		if len(*project.RepoPath) > bestLength {
			best = project
			bestLength = len(*project.RepoPath)
		}
	}
	return best
}

// pathContains reports whether child is parent or lives beneath it, comparing
// path components rather than raw strings.
func pathContains(parent, child string) bool {
	relative, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	if relative == "." {
		return true
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return false
	}
	return !filepath.IsAbs(relative)
}

// resolveCwd returns the canonical absolute current directory. A directory
// that cannot be resolved falls back to its absolute form.
func resolveCwd() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	absolute, err := filepath.Abs(cwd)
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(absolute); err == nil {
		return resolved, nil
	}
	return absolute, nil
}
