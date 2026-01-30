package detector

import (
	"path/filepath"
	"regexp"
	"strings"

	"repoctr/pkg/models"
)

type goDetector struct{}

func NewGoDetector() Detector {
	return &goDetector{}
}

func (d *goDetector) Name() string {
	return "Go"
}

func (d *goDetector) RuntimeType() models.RuntimeType {
	return models.RuntimeGo
}

func (d *goDetector) ManifestFiles() []string {
	return []string{"go.mod"}
}

func (d *goDetector) Detect(manifestPath string, content []byte) (*models.Project, error) {
	if filepath.Base(manifestPath) != "go.mod" {
		return nil, nil
	}

	contentStr := string(content)

	// Check for module declaration
	if !strings.Contains(contentStr, "module ") {
		return nil, nil
	}

	// Extract module name
	name := ""
	moduleRe := regexp.MustCompile(`module\s+(\S+)`)
	if matches := moduleRe.FindStringSubmatch(contentStr); len(matches) > 1 {
		name = matches[1]
		// Use the last part of the module path as name
		parts := strings.Split(name, "/")
		name = parts[len(parts)-1]
	}

	// Extract Go version
	version := ""
	versionRe := regexp.MustCompile(`go\s+(\d+\.\d+)`)
	if matches := versionRe.FindStringSubmatch(contentStr); len(matches) > 1 {
		version = matches[1]
	}

	dir := filepath.Dir(manifestPath)
	if name == "" {
		name = filepath.Base(dir)
	}

	return &models.Project{
		Name:           name,
		Path:           dir,
		Runtime:        models.Runtime{Type: models.RuntimeGo, Version: version},
		ManifestFile:   "go.mod",
		SourcePaths:    []string{"."},
		SrcIgnorePaths: []string{"vendor"},
	}, nil
}
