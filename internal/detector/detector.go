package detector

import (
	"repoctr/pkg/models"
)

// Detector defines the interface for project detection.
type Detector interface {
	// Name returns the detector name for logging/debugging.
	Name() string

	// RuntimeType returns the runtime type this detector handles.
	RuntimeType() models.RuntimeType

	// ManifestFiles returns the list of manifest file patterns to look for.
	ManifestFiles() []string

	// Detect checks if a manifest file represents a project and extracts info.
	// Returns the project if detected, nil if not applicable.
	Detect(manifestPath string, content []byte) (*models.Project, error)
}

// Registry holds all registered detectors.
type Registry struct {
	detectors []Detector
}

// NewRegistry creates a new detector registry with all built-in detectors.
func NewRegistry() *Registry {
	return &Registry{
		detectors: []Detector{
			NewDotNetDetector(),
			NewPythonDetector(),
			NewGoDetector(),
			NewJavaDetector(),
			NewJavaScriptDetector(),
			NewDartDetector(),
			NewCppDetector(),
			NewRustDetector(),
		},
	}
}

// Detectors returns all registered detectors.
func (r *Registry) Detectors() []Detector {
	return r.detectors
}

// GetManifestPatterns returns all manifest file patterns across all detectors.
func (r *Registry) GetManifestPatterns() []string {
	patterns := make([]string, 0)
	for _, d := range r.detectors {
		patterns = append(patterns, d.ManifestFiles()...)
	}
	return patterns
}

// DetectProject tries all detectors for a given manifest file.
func (r *Registry) DetectProject(manifestPath string, content []byte) (*models.Project, error) {
	for _, d := range r.detectors {
		project, err := d.Detect(manifestPath, content)
		if err != nil {
			return nil, err
		}
		if project != nil {
			return project, nil
		}
	}
	return nil, nil
}
