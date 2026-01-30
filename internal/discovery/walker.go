package discovery

import (
	"os"
	"path/filepath"
	"strings"

	"repoctr/internal/detector"
	"repoctr/internal/ignore"
	"repoctr/pkg/models"
)

// Walker handles recursive directory traversal for project discovery.
type Walker struct {
	registry *detector.Registry
	matcher  *ignore.Matcher
	rootDir  string
}

// NewWalker creates a new walker for the given root directory.
func NewWalker(rootDir string, registry *detector.Registry) (*Walker, error) {
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, err
	}

	matcher, err := ignore.NewMatcher(absRoot)
	if err != nil {
		return nil, err
	}

	return &Walker{
		registry: registry,
		matcher:  matcher,
		rootDir:  absRoot,
	}, nil
}

// Discover walks the directory tree and returns all discovered projects.
func (w *Walker) Discover() ([]*models.Project, error) {
	var projects []*models.Project
	manifestPatterns := w.registry.GetManifestPatterns()

	err := filepath.WalkDir(w.rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip inaccessible paths
		}

		// Skip ignored directories
		if d.IsDir() {
			if w.matcher.ShouldIgnore(path) {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if this file matches any manifest pattern
		filename := d.Name()
		if !w.matchesManifest(filename, manifestPatterns) {
			return nil
		}

		// Skip ignored files
		if w.matcher.ShouldIgnoreFile(path) {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip unreadable files
		}

		// Try to detect project
		project, err := w.registry.DetectProject(path, content)
		if err != nil {
			return nil // Skip detection errors
		}

		if project != nil {
			// Make path relative to root
			relPath, err := filepath.Rel(w.rootDir, project.Path)
			if err == nil {
				project.Path = relPath
			}
			projects = append(projects, project)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return projects, nil
}

// matchesManifest checks if a filename matches any manifest pattern.
func (w *Walker) matchesManifest(filename string, patterns []string) bool {
	for _, pattern := range patterns {
		// Check for exact match
		if pattern == filename {
			return true
		}

		// Check for glob pattern match
		if strings.Contains(pattern, "*") {
			matched, err := filepath.Match(pattern, filename)
			if err == nil && matched {
				return true
			}
		}
	}
	return false
}
