package models

// RepoCtrConfig represents the user configuration in .repoctrconfig.yaml.
type RepoCtrConfig struct {
	GlobalExcludes   []string                   `yaml:"global-excludes,omitempty"`
	ProjectOverrides map[string]ProjectOverride `yaml:"project-overrides,omitempty"`
}

// ProjectOverride contains project-specific configuration overrides.
type ProjectOverride struct {
	ExcludePatterns []string `yaml:"exclude-patterns,omitempty"`
	SrcIgnorePaths  []string `yaml:"src-ignore-paths,omitempty"`
	SourcePaths     []string `yaml:"source-paths,omitempty"`
}
