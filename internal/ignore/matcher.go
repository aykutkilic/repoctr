package ignore

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Matcher handles gitignore patterns and custom ignore rules.
type Matcher struct {
	rootDir        string
	defaultIgnores map[string]bool
	gitignoreRules []gitignoreRule
}

type gitignoreRule struct {
	pattern  string
	negate   bool
	dirOnly  bool
	anchored bool
}

// DefaultIgnorePatterns contains patterns that should always be ignored.
var DefaultIgnorePatterns = []string{
	// Version control
	".git",
	".svn",
	".hg",
	// Dependencies/packages
	"node_modules",
	"vendor",
	// Python build/cache
	"__pycache__",
	".tox",
	".nox",
	".eggs",
	"*.egg-info",
	"__pypackages__",
	// Python virtual environments (various naming conventions)
	"venv",
	".venv",
	"env",
	".env",
	"ENV",
	"virtualenv",
	".virtualenv",
	"pythonenv",
	".pythonenv",
	".conda",
	// Build outputs
	"target",
	"build",
	"dist",
	".gradle",
	// IDE/editor
	".idea",
	".vscode",
	".vs",
	// Compiled outputs
	"bin",
	"obj",
	// OS files
	".DS_Store",
	"Thumbs.db",
}

// DefaultIgnoreExtensions contains file extensions to ignore.
var DefaultIgnoreExtensions = []string{
	".pyc",
	".pyo",
	".class",
	".o",
	".a",
	".so",
	".dylib",
}

// NewMatcher creates a new ignore matcher for the given root directory.
func NewMatcher(rootDir string) (*Matcher, error) {
	m := &Matcher{
		rootDir:        rootDir,
		defaultIgnores: make(map[string]bool),
	}

	// Build default ignore set
	for _, pattern := range DefaultIgnorePatterns {
		m.defaultIgnores[pattern] = true
	}

	// Load .gitignore if it exists
	gitignorePath := filepath.Join(rootDir, ".gitignore")
	if rules, err := parseGitignore(gitignorePath); err == nil {
		m.gitignoreRules = rules
	}

	return m, nil
}

// parseGitignore reads and parses a .gitignore file.
func parseGitignore(path string) ([]gitignoreRule, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var rules []gitignoreRule
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		rule := gitignoreRule{}

		// Check for negation
		if strings.HasPrefix(line, "!") {
			rule.negate = true
			line = line[1:]
		}

		// Check for directory only
		if strings.HasSuffix(line, "/") {
			rule.dirOnly = true
			line = strings.TrimSuffix(line, "/")
		}

		// Check if anchored (contains / not at end)
		if strings.Contains(line, "/") {
			rule.anchored = true
		}

		rule.pattern = line
		rules = append(rules, rule)
	}

	return rules, scanner.Err()
}

// ShouldIgnore checks if a path should be ignored.
func (m *Matcher) ShouldIgnore(path string) bool {
	// Get relative path from root
	relPath, err := filepath.Rel(m.rootDir, path)
	if err != nil {
		relPath = path
	}

	// Normalize to forward slashes for matching
	relPath = filepath.ToSlash(relPath)

	// Check if it's a directory
	info, err := os.Stat(path)
	isDir := err == nil && info.IsDir()

	// Check basename against default patterns
	base := filepath.Base(path)
	if m.defaultIgnores[base] {
		return true
	}

	// Check file extensions
	if !isDir {
		ext := strings.ToLower(filepath.Ext(path))
		for _, ignoreExt := range DefaultIgnoreExtensions {
			if ext == ignoreExt {
				return true
			}
		}
	}

	// Check gitignore rules
	if m.matchGitignore(relPath, isDir) {
		return true
	}

	return false
}

// ShouldIgnoreFile checks if a file path should be ignored (not directory check).
func (m *Matcher) ShouldIgnoreFile(path string) bool {
	relPath, err := filepath.Rel(m.rootDir, path)
	if err != nil {
		relPath = path
	}

	// Normalize to forward slashes for matching
	relPath = filepath.ToSlash(relPath)

	// Check basename against default patterns
	base := filepath.Base(path)
	if m.defaultIgnores[base] {
		return true
	}

	// Check file extensions
	ext := strings.ToLower(filepath.Ext(path))
	for _, ignoreExt := range DefaultIgnoreExtensions {
		if ext == ignoreExt {
			return true
		}
	}

	// Check gitignore rules
	if m.matchGitignore(relPath, false) {
		return true
	}

	return false
}

// matchGitignore checks if a path matches any gitignore rule.
func (m *Matcher) matchGitignore(relPath string, isDir bool) bool {
	ignored := false

	for _, rule := range m.gitignoreRules {
		// Skip directory-only rules for files
		if rule.dirOnly && !isDir {
			continue
		}

		matched := false

		if rule.anchored {
			// Anchored patterns match from root
			matched = matchPattern(rule.pattern, relPath)
		} else {
			// Non-anchored patterns match any path component
			matched = matchPattern(rule.pattern, relPath) ||
				matchPattern(rule.pattern, filepath.Base(relPath))
		}

		if matched {
			ignored = !rule.negate
		}
	}

	return ignored
}

// matchPattern performs simple glob matching.
func matchPattern(pattern, path string) bool {
	// Handle ** for recursive matching
	if strings.Contains(pattern, "**") {
		parts := strings.Split(pattern, "**")
		if len(parts) == 2 {
			prefix := strings.TrimSuffix(parts[0], "/")
			suffix := strings.TrimPrefix(parts[1], "/")

			if prefix == "" && suffix == "" {
				return true
			}
			if prefix != "" && !strings.HasPrefix(path, prefix) {
				return false
			}
			if suffix != "" && !strings.HasSuffix(path, suffix) {
				return false
			}
			return true
		}
	}

	// Try exact match
	if pattern == path {
		return true
	}

	// Try filepath.Match for glob patterns
	if matched, err := filepath.Match(pattern, path); err == nil && matched {
		return true
	}

	// Try matching against basename
	if matched, err := filepath.Match(pattern, filepath.Base(path)); err == nil && matched {
		return true
	}

	// Try prefix match for directory patterns
	if strings.HasPrefix(path, pattern+"/") {
		return true
	}

	return false
}
