package detector

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"repoctr/pkg/models"
)

type pythonDetector struct{}

func NewPythonDetector() Detector {
	return &pythonDetector{}
}

func (d *pythonDetector) Name() string {
	return "Python"
}

func (d *pythonDetector) RuntimeType() models.RuntimeType {
	return models.RuntimePython
}

func (d *pythonDetector) ManifestFiles() []string {
	return []string{"pyproject.toml", "setup.py", "requirements.txt"}
}

func (d *pythonDetector) Detect(manifestPath string, content []byte) (*models.Project, error) {
	filename := filepath.Base(manifestPath)

	switch filename {
	case "pyproject.toml":
		return d.detectPyprojectToml(manifestPath, content)
	case "setup.py":
		return d.detectSetupPy(manifestPath, content)
	case "requirements.txt":
		return d.detectRequirementsTxt(manifestPath, content)
	}

	return nil, nil
}

// pyprojectToml represents the structure of a pyproject.toml file.
type pyprojectToml struct {
	Project struct {
		Name           string `toml:"name"`
		RequiresPython string `toml:"requires-python"`
	} `toml:"project"`
	Tool struct {
		Poetry struct {
			Name   string `toml:"name"`
			Python string `toml:"python"`
		} `toml:"poetry"`
	} `toml:"tool"`
	BuildSystem struct {
		Requires []string `toml:"requires"`
	} `toml:"build-system"`
}

func (d *pythonDetector) detectPyprojectToml(manifestPath string, content []byte) (*models.Project, error) {
	var pyproj pyprojectToml
	if _, err := toml.Decode(string(content), &pyproj); err != nil {
		// If TOML parsing fails, still detect as Python project
		return d.createProject(manifestPath, "", ""), nil
	}

	// Determine project name
	name := pyproj.Project.Name
	if name == "" {
		name = pyproj.Tool.Poetry.Name
	}

	// Determine Python version
	version := pyproj.Project.RequiresPython
	if version == "" {
		version = pyproj.Tool.Poetry.Python
	}
	version = cleanPythonVersion(version)

	return d.createProject(manifestPath, name, version), nil
}

func (d *pythonDetector) detectSetupPy(manifestPath string, content []byte) (*models.Project, error) {
	contentStr := string(content)

	// Check if it's a valid setup.py
	if !strings.Contains(contentStr, "setup(") && !strings.Contains(contentStr, "setup (") {
		return nil, nil
	}

	// Try to extract name
	name := ""
	nameRe := regexp.MustCompile(`name\s*=\s*["']([^"']+)["']`)
	if matches := nameRe.FindStringSubmatch(contentStr); len(matches) > 1 {
		name = matches[1]
	}

	// Try to extract python_requires
	version := ""
	versionRe := regexp.MustCompile(`python_requires\s*=\s*["']([^"']+)["']`)
	if matches := versionRe.FindStringSubmatch(contentStr); len(matches) > 1 {
		version = cleanPythonVersion(matches[1])
	}

	return d.createProject(manifestPath, name, version), nil
}

func (d *pythonDetector) detectRequirementsTxt(manifestPath string, content []byte) (*models.Project, error) {
	// requirements.txt is a valid Python project indicator
	// but provides no name or version info
	return d.createProject(manifestPath, "", ""), nil
}

func (d *pythonDetector) createProject(manifestPath, name, version string) *models.Project {
	dir := filepath.Dir(manifestPath)
	if name == "" {
		name = filepath.Base(dir)
	}

	return &models.Project{
		Name:         name,
		Path:         dir,
		Runtime:      models.Runtime{Type: models.RuntimePython, Version: version},
		ManifestFile: filepath.Base(manifestPath),
		SourcePaths:  []string{"src", "."},
	}
}

// cleanPythonVersion extracts version from requirement specifiers.
// Examples: ">=3.8" -> "3.8+", "^3.9" -> "3.9+", ">=3.8,<4" -> "3.8+"
func cleanPythonVersion(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}

	// Extract the first version number
	re := regexp.MustCompile(`(\d+\.?\d*\.?\d*)`)
	if matches := re.FindStringSubmatch(v); len(matches) > 1 {
		version := matches[1]
		// Add + suffix if it's a minimum version
		if strings.HasPrefix(v, ">=") || strings.HasPrefix(v, "^") || strings.HasPrefix(v, ">") {
			return version + "+"
		}
		return version
	}

	return v
}
