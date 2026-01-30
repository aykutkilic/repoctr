package discovery

import (
	"path/filepath"
	"sort"
	"strings"

	"repoctr/pkg/models"
)

// HierarchyBuilder builds a nested project tree from a flat list.
type HierarchyBuilder struct{}

// NewHierarchyBuilder creates a new hierarchy builder.
func NewHierarchyBuilder() *HierarchyBuilder {
	return &HierarchyBuilder{}
}

// Build creates a hierarchical project tree from a flat list of projects.
// Projects are nested based on their filesystem paths.
func (b *HierarchyBuilder) Build(projects []*models.Project) []*models.Project {
	if len(projects) == 0 {
		return nil
	}

	// Sort projects by path depth (shallowest first)
	sorted := make([]*models.Project, len(projects))
	copy(sorted, projects)
	sort.Slice(sorted, func(i, j int) bool {
		depthI := strings.Count(sorted[i].Path, string(filepath.Separator))
		depthJ := strings.Count(sorted[j].Path, string(filepath.Separator))
		if depthI != depthJ {
			return depthI < depthJ
		}
		return sorted[i].Path < sorted[j].Path
	})

	// Build path map for quick parent lookup
	pathMap := make(map[string]*models.Project)
	var roots []*models.Project

	for _, project := range sorted {
		pathMap[project.Path] = project

		// Find nearest ancestor
		parent := b.findNearestAncestor(project.Path, pathMap)
		if parent != nil {
			parent.Children = append(parent.Children, project)
		} else {
			roots = append(roots, project)
		}
	}

	return roots
}

// findNearestAncestor finds the closest ancestor project in the path map.
func (b *HierarchyBuilder) findNearestAncestor(path string, pathMap map[string]*models.Project) *models.Project {
	// Walk up the directory tree looking for a parent project
	current := filepath.Dir(path)

	for current != "." && current != "/" && current != "" {
		if parent, ok := pathMap[current]; ok {
			return parent
		}
		current = filepath.Dir(current)
	}

	// Also check for root path "."
	if path != "." {
		if parent, ok := pathMap["."]; ok {
			return parent
		}
	}

	return nil
}

// Flatten converts a hierarchical project tree back to a flat list.
func (b *HierarchyBuilder) Flatten(roots []*models.Project) []*models.Project {
	var result []*models.Project

	var flatten func(projects []*models.Project)
	flatten = func(projects []*models.Project) {
		for _, p := range projects {
			result = append(result, p)
			if len(p.Children) > 0 {
				flatten(p.Children)
			}
		}
	}

	flatten(roots)
	return result
}
