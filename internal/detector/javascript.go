package detector

import (
	"encoding/json"
	"os"
	"path/filepath"

	"repoctr/pkg/models"
)

type javascriptDetector struct{}

func NewJavaScriptDetector() Detector {
	return &javascriptDetector{}
}

func (d *javascriptDetector) Name() string {
	return "JavaScript"
}

func (d *javascriptDetector) RuntimeType() models.RuntimeType {
	return models.RuntimeJavaScript
}

func (d *javascriptDetector) ManifestFiles() []string {
	return []string{"package.json"}
}

func (d *javascriptDetector) Detect(manifestPath string, content []byte) (*models.Project, error) {
	if filepath.Base(manifestPath) != "package.json" {
		return nil, nil
	}

	var pkg packageJSON
	if err := json.Unmarshal(content, &pkg); err != nil {
		// If JSON parsing fails, still detect as JS project
		return d.createProject(manifestPath, "", "", false), nil
	}

	// Determine if TypeScript
	isTypeScript := d.isTypeScriptProject(manifestPath, pkg)

	// Get Node.js version from engines
	nodeVersion := ""
	if pkg.Engines.Node != "" {
		nodeVersion = pkg.Engines.Node
	}

	return d.createProject(manifestPath, pkg.Name, nodeVersion, isTypeScript), nil
}

// packageJSON represents the structure of a package.json file.
type packageJSON struct {
	Name            string            `json:"name"`
	Engines         engines           `json:"engines"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

type engines struct {
	Node string `json:"node"`
}

func (d *javascriptDetector) isTypeScriptProject(manifestPath string, pkg packageJSON) bool {
	dir := filepath.Dir(manifestPath)

	// Check for tsconfig.json
	if _, err := os.Stat(filepath.Join(dir, "tsconfig.json")); err == nil {
		return true
	}

	// Check for typescript in dependencies
	if _, ok := pkg.Dependencies["typescript"]; ok {
		return true
	}
	if _, ok := pkg.DevDependencies["typescript"]; ok {
		return true
	}

	return false
}

func (d *javascriptDetector) createProject(manifestPath, name, version string, isTypeScript bool) *models.Project {
	dir := filepath.Dir(manifestPath)
	if name == "" {
		name = filepath.Base(dir)
	}

	runtimeType := models.RuntimeJavaScript
	if isTypeScript {
		runtimeType = models.RuntimeTypeScript
	}

	return &models.Project{
		Name:           name,
		Path:           dir,
		Runtime:        models.Runtime{Type: runtimeType, Version: version},
		ManifestFile:   "package.json",
		SourcePaths:    []string{"src", "lib", "."},
		SrcIgnorePaths: []string{"node_modules", "dist", "build"},
	}
}
