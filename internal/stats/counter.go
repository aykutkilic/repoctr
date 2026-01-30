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

			// Skip non-source files (only count files for this project's runtime)
			if !isSourceFile(path, project.Runtime.Type) {
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

// sourceExtensionsByRuntime maps each RuntimeType to its language-specific source file extensions.
// LOC is calculated only on source files relevant to the detected project type.
var sourceExtensionsByRuntime = map[models.RuntimeType]map[string]bool{
	models.RuntimeGo: {
		".go": true,
	},
	models.RuntimePython: {
		".py": true, ".pyw": true, ".pyi": true,
	},
	models.RuntimeJavaScript: {
		".js": true, ".jsx": true, ".mjs": true, ".cjs": true,
	},
	models.RuntimeTypeScript: {
		".ts": true, ".tsx": true, ".js": true, ".jsx": true, ".mjs": true, ".cjs": true,
	},
	models.RuntimeJava: {
		".java": true, ".kt": true, ".kts": true, ".scala": true,
	},
	models.RuntimeDotNet: {
		".cs": true, ".fs": true, ".vb": true,
	},
	models.RuntimeRust: {
		".rs": true,
	},
	models.RuntimeDart: {
		".dart": true,
	},
	models.RuntimeCpp: {
		".c": true, ".h": true, ".cpp": true, ".cc": true, ".cxx": true,
		".hpp": true, ".hh": true, ".hxx": true,
	},
}

// isSourceFile checks if a file is a source code file for the given runtime type.
func isSourceFile(path string, runtimeType models.RuntimeType) bool {
	ext := strings.ToLower(filepath.Ext(path))

	// Get extensions for the specific runtime
	if exts, ok := sourceExtensionsByRuntime[runtimeType]; ok {
		return exts[ext]
	}

	// Fallback: if runtime type is unknown, accept no files
	// This ensures we don't accidentally count unrelated files
	return false
}
