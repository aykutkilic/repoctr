package config

import (
	"repoctr/pkg/models"
)

// MergeProjects combines discovered projects with existing projects and applies
// overrides from the configuration. It performs a non-destructive merge that
// preserves user customizations while updating auto-detected fields.
func MergeProjects(
	discovered []*models.Project,
	existing []*models.Project,
	cfg *models.RepoCtrConfig,
) []*models.Project {
	// Build a map of existing projects by path for fast lookup
	existingMap := buildProjectMap(existing)

	var result []*models.Project

	// Process discovered projects
	for _, discoveredProj := range discovered {
		// Check if this project already exists
		if existingProj, found := existingMap[discoveredProj.Path]; found {
			// Merge discovered into existing
			merged := mergeProject(existingProj, discoveredProj)
			applyConfigOverrides(merged, cfg)
			result = append(result, merged)
			delete(existingMap, discoveredProj.Path)
		} else {
			// New project - just apply config overrides
			applyConfigOverrides(discoveredProj, cfg)
			result = append(result, discoveredProj)
		}
	}

	// Add any existing projects that weren't re-discovered
	// (but still apply config overrides)
	for _, existingProj := range existingMap {
		applyConfigOverrides(existingProj, cfg)
		result = append(result, existingProj)
	}

	return result
}

// buildProjectMap creates a map of projects by their path for quick lookup.
func buildProjectMap(projects []*models.Project) map[string]*models.Project {
	m := make(map[string]*models.Project)
	for _, p := range projects {
		m[p.Path] = p
	}
	return m
}

// mergeProject merges discovered project info into an existing project,
// preserving user-customized fields while updating auto-detected ones.
func mergeProject(existing, discovered *models.Project) *models.Project {
	result := &models.Project{
		// Keep existing values where user might have customized
		Name:           discovered.Name, // Use discovered name
		Path:           existing.Path,   // Path is the primary key
		Runtime:        discovered.Runtime,
		ManifestFile:   discovered.ManifestFile,
		SourcePaths:    discovered.SourcePaths,
		ExcludePatterns: existing.ExcludePatterns, // Preserve user excludes
		Children:       discovered.Children,       // Use discovered hierarchy
	}

	// For src-ignore-paths, if user has set them, keep them; otherwise use discovered
	if len(existing.SrcIgnorePaths) > 0 {
		result.SrcIgnorePaths = existing.SrcIgnorePaths
	} else {
		result.SrcIgnorePaths = discovered.SrcIgnorePaths
	}

	return result
}

// applyConfigOverrides applies configuration overrides from .repoctrconfig.yaml
// to a project.
func applyConfigOverrides(project *models.Project, cfg *models.RepoCtrConfig) {
	if cfg == nil || cfg.ProjectOverrides == nil {
		return
	}

	// Check for project-specific overrides by path
	if override, found := cfg.ProjectOverrides[project.Path]; found {
		// Apply exclude patterns override if provided
		if len(override.ExcludePatterns) > 0 {
			project.ExcludePatterns = override.ExcludePatterns
		}

		// Apply src-ignore-paths override if provided
		if len(override.SrcIgnorePaths) > 0 {
			project.SrcIgnorePaths = override.SrcIgnorePaths
		}

		// Apply source-paths override if provided
		if len(override.SourcePaths) > 0 {
			project.SourcePaths = override.SourcePaths
		}
	}
}
