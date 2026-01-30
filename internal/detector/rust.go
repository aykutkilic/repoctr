package detector

import (
	"path/filepath"

	"github.com/BurntSushi/toml"
	"repoctr/pkg/models"
)

type rustDetector struct{}

func NewRustDetector() Detector {
	return &rustDetector{}
}

func (d *rustDetector) Name() string {
	return "Rust"
}

func (d *rustDetector) RuntimeType() models.RuntimeType {
	return models.RuntimeRust
}

func (d *rustDetector) ManifestFiles() []string {
	return []string{"Cargo.toml"}
}

func (d *rustDetector) Detect(manifestPath string, content []byte) (*models.Project, error) {
	if filepath.Base(manifestPath) != "Cargo.toml" {
		return nil, nil
	}

	var cargo cargoToml
	if _, err := toml.Decode(string(content), &cargo); err != nil {
		// If TOML parsing fails, still detect as Rust project
		return d.createProject(manifestPath, "", ""), nil
	}

	// Get rust version
	version := cargo.Package.RustVersion
	if version == "" {
		version = cargo.Package.Edition
	}

	return d.createProject(manifestPath, cargo.Package.Name, version), nil
}

// cargoToml represents the structure of a Cargo.toml file.
type cargoToml struct {
	Package struct {
		Name        string `toml:"name"`
		RustVersion string `toml:"rust-version"`
		Edition     string `toml:"edition"`
	} `toml:"package"`
	Workspace struct {
		Members []string `toml:"members"`
	} `toml:"workspace"`
}

func (d *rustDetector) createProject(manifestPath, name, version string) *models.Project {
	dir := filepath.Dir(manifestPath)
	if name == "" {
		name = filepath.Base(dir)
	}

	return &models.Project{
		Name:           name,
		Path:           dir,
		Runtime:        models.Runtime{Type: models.RuntimeRust, Version: version},
		ManifestFile:   "Cargo.toml",
		SourcePaths:    []string{"src"},
		SrcIgnorePaths: []string{"target"},
	}
}
