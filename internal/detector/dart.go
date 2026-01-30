package detector

import (
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
	"repoctr/pkg/models"
)

type dartDetector struct{}

func NewDartDetector() Detector {
	return &dartDetector{}
}

func (d *dartDetector) Name() string {
	return "Dart"
}

func (d *dartDetector) RuntimeType() models.RuntimeType {
	return models.RuntimeDart
}

func (d *dartDetector) ManifestFiles() []string {
	return []string{"pubspec.yaml"}
}

func (d *dartDetector) Detect(manifestPath string, content []byte) (*models.Project, error) {
	if filepath.Base(manifestPath) != "pubspec.yaml" {
		return nil, nil
	}

	var pubspec pubspecYaml
	if err := yaml.Unmarshal(content, &pubspec); err != nil {
		// If YAML parsing fails, still detect as Dart project
		return d.createProject(manifestPath, "", ""), nil
	}

	// Extract SDK version from environment
	sdkVersion := ""
	if pubspec.Environment.SDK != "" {
		sdkVersion = cleanDartVersion(pubspec.Environment.SDK)
	}

	return d.createProject(manifestPath, pubspec.Name, sdkVersion), nil
}

// pubspecYaml represents the structure of a pubspec.yaml file.
type pubspecYaml struct {
	Name        string `yaml:"name"`
	Environment struct {
		SDK     string `yaml:"sdk"`
		Flutter string `yaml:"flutter"`
	} `yaml:"environment"`
}

func (d *dartDetector) createProject(manifestPath, name, version string) *models.Project {
	dir := filepath.Dir(manifestPath)
	if name == "" {
		name = filepath.Base(dir)
	}

	return &models.Project{
		Name:           name,
		Path:           dir,
		Runtime:        models.Runtime{Type: models.RuntimeDart, Version: version},
		ManifestFile:   "pubspec.yaml",
		SourcePaths:    []string{"lib", "bin"},
		SrcIgnorePaths: []string{".dart_tool", "build"},
	}
}

// cleanDartVersion extracts version from SDK constraint.
// Examples: ">=3.0.0 <4.0.0" -> "3.0.0+", "^3.2.0" -> "3.2.0+"
func cleanDartVersion(v string) string {
	re := regexp.MustCompile(`(\d+\.\d+\.?\d*)`)
	if matches := re.FindStringSubmatch(v); len(matches) > 1 {
		return matches[1] + "+"
	}
	return v
}
