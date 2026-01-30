package stats

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"repoctr/internal/ignore"
	"repoctr/pkg/models"
)

// Counter calculates LOC statistics for projects.
type Counter struct {
	rootDir string
	matcher *ignore.Matcher
}

// NewCounter creates a new stats counter.
func NewCounter(rootDir string) (*Counter, error) {
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, err
	}

	matcher, err := ignore.NewMatcher(absRoot)
	if err != nil {
		return nil, err
	}

	return &Counter{
		rootDir: absRoot,
		matcher: matcher,
	}, nil
}

// CountProject calculates statistics for a single project.
func (c *Counter) CountProject(project *models.Project) (*models.ProjectStats, error) {
	stats := &models.ProjectStats{
		Project:      project,
		LargestFiles: make([]models.FileStats, 0, 5),
	}

	// Build the full project path
	projectPath := filepath.Join(c.rootDir, project.Path)

	// Track all file stats for finding largest
	var allFiles []models.FileStats
	folderSet := make(map[string]bool)

	// Process each source path
	for _, srcPath := range project.SourcePaths {
		fullPath := filepath.Join(projectPath, srcPath)

		// Check if path exists
		info, err := os.Stat(fullPath)
		if err != nil {
			continue // Skip non-existent paths
		}

		if !info.IsDir() {
			// Single file
			fileStats, err := c.countFile(fullPath)
			if err == nil {
				c.addFileStats(stats, fileStats)
				allFiles = append(allFiles, *fileStats)
			}
			continue
		}

		// Walk directory
		err = filepath.WalkDir(fullPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			// Get relative path from project root for ignore checking
			relPath, _ := filepath.Rel(projectPath, path)

			// Check if should be ignored
			if d.IsDir() {
				// Check against project-specific ignore paths
				for _, ignorePath := range project.SrcIgnorePaths {
					if relPath == ignorePath || strings.HasPrefix(relPath, ignorePath+string(filepath.Separator)) {
						return filepath.SkipDir
					}
				}

				if c.matcher.ShouldIgnore(path) {
					return filepath.SkipDir
				}
				folderSet[path] = true
				return nil
			}

			// Skip non-source files
			if !isSourceFile(path) {
				return nil
			}

			// Skip ignored files
			if c.matcher.ShouldIgnoreFile(path) {
				return nil
			}

			fileStats, err := c.countFile(path)
			if err == nil {
				c.addFileStats(stats, fileStats)
				allFiles = append(allFiles, *fileStats)
			}

			return nil
		})
		if err != nil {
			continue
		}
	}

	stats.TotalFolders = len(folderSet)

	// Find top 5 largest files
	sort.Slice(allFiles, func(i, j int) bool {
		return allFiles[i].Lines > allFiles[j].Lines
	})

	limit := 5
	if len(allFiles) < limit {
		limit = len(allFiles)
	}
	stats.LargestFiles = allFiles[:limit]

	return stats, nil
}

// CountHierarchy calculates statistics for a project hierarchy.
func (c *Counter) CountHierarchy(projects []*models.Project) ([]*models.ProjectStats, error) {
	var results []*models.ProjectStats

	for _, project := range projects {
		stats, err := c.CountProject(project)
		if err != nil {
			continue
		}

		// Recursively count children
		if len(project.Children) > 0 {
			childStats, err := c.CountHierarchy(project.Children)
			if err == nil {
				stats.Children = childStats
			}
		}

		results = append(results, stats)
	}

	return results, nil
}

func (c *Counter) countFile(path string) (*models.FileStats, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	stats := &models.FileStats{
		Path: path,
		Size: info.Size(),
	}

	scanner := bufio.NewScanner(file)
	// Handle long lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		stats.Lines++

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			stats.BlankLines++
		} else {
			stats.CodeLines++
		}
	}

	return stats, scanner.Err()
}

func (c *Counter) addFileStats(projectStats *models.ProjectStats, fileStats *models.FileStats) {
	projectStats.TotalFiles++
	projectStats.TotalLines += fileStats.Lines
	projectStats.BlankLines += fileStats.BlankLines
	projectStats.CodeLines += fileStats.CodeLines
	projectStats.TotalSize += fileStats.Size
}

// isSourceFile checks if a file is a source code file based on extension.
func isSourceFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	sourceExts := map[string]bool{
		// Go
		".go": true,
		// Python
		".py": true, ".pyw": true, ".pyi": true,
		// JavaScript/TypeScript
		".js": true, ".jsx": true, ".ts": true, ".tsx": true, ".mjs": true, ".cjs": true,
		// Java
		".java": true, ".kt": true, ".kts": true, ".scala": true,
		// C/C++
		".c": true, ".h": true, ".cpp": true, ".cc": true, ".cxx": true,
		".hpp": true, ".hh": true, ".hxx": true,
		// C#
		".cs": true, ".fs": true, ".vb": true,
		// Rust
		".rs": true,
		// Dart
		".dart": true,
		// Ruby
		".rb": true, ".rake": true,
		// PHP
		".php": true,
		// Swift
		".swift": true,
		// Objective-C
		".m": true, ".mm": true,
		// Shell
		".sh": true, ".bash": true, ".zsh": true,
		// Config/Data
		".json": true, ".yaml": true, ".yml": true, ".toml": true,
		".xml": true, ".html": true, ".css": true, ".scss": true, ".less": true,
		// SQL
		".sql": true,
		// Markdown
		".md": true, ".markdown": true,
	}
	return sourceExts[ext]
}
