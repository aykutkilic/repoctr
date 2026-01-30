package models

// RuntimeType represents the programming language/runtime of a project.
type RuntimeType string

const (
	RuntimeDotNet     RuntimeType = ".NET"
	RuntimePython     RuntimeType = "Python"
	RuntimeGo         RuntimeType = "Go"
	RuntimeJava       RuntimeType = "Java"
	RuntimeTypeScript RuntimeType = "TypeScript"
	RuntimeJavaScript RuntimeType = "JavaScript"
	RuntimeDart       RuntimeType = "Dart"
	RuntimeCpp        RuntimeType = "C/C++"
	RuntimeRust       RuntimeType = "Rust"
)

// Runtime describes the language runtime and version for a project.
type Runtime struct {
	Type    RuntimeType `yaml:"type"`
	Version string      `yaml:"version,omitempty"`
}

// Project represents a discovered project in the repository.
type Project struct {
	Name           string     `yaml:"name"`
	Path           string     `yaml:"path"`
	Runtime        Runtime    `yaml:"runtime"`
	ManifestFile   string     `yaml:"manifest-file"`
	SourcePaths    []string   `yaml:"source-paths"`
	SrcIgnorePaths []string   `yaml:"src-ignore-paths,omitempty"`
	Children       []*Project `yaml:"children,omitempty"`
}

// ProjectsConfig is the root structure for projects.yaml.
type ProjectsConfig struct {
	Projects []*Project `yaml:"projects"`
}
