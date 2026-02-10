package detector

import (
	"encoding/xml"
	"path/filepath"
	"regexp"
	"strings"

	"repoctr/pkg/models"
)

type dotNetDetector struct{}

func NewDotNetDetector() Detector {
	return &dotNetDetector{}
}

func (d *dotNetDetector) Name() string {
	return "DotNet"
}

func (d *dotNetDetector) RuntimeType() models.RuntimeType {
	return models.RuntimeDotNet
}

func (d *dotNetDetector) ManifestFiles() []string {
	return []string{"*.csproj", "*.sln", "*.fsproj", "*.vbproj"}
}

func (d *dotNetDetector) Detect(manifestPath string, content []byte) (*models.Project, error) {
	ext := strings.ToLower(filepath.Ext(manifestPath))

	switch ext {
	case ".csproj", ".fsproj", ".vbproj":
		return d.detectProjectFile(manifestPath, content)
	case ".sln":
		return d.detectSolutionFile(manifestPath, content)
	}

	return nil, nil
}

// csprojFile represents the structure of a .csproj XML file.
type csprojFile struct {
	XMLName        xml.Name        `xml:"Project"`
	PropertyGroups []propertyGroup `xml:"PropertyGroup"`
}

type propertyGroup struct {
	TargetFramework  string `xml:"TargetFramework"`
	TargetFrameworks string `xml:"TargetFrameworks"`
}

func (d *dotNetDetector) detectProjectFile(manifestPath string, content []byte) (*models.Project, error) {
	// Check if this is a .NET project file
	if !strings.Contains(string(content), "<Project") {
		return nil, nil
	}

	var proj csprojFile
	if err := xml.Unmarshal(content, &proj); err != nil {
		// If XML parsing fails, still detect as .NET project but without version
		return d.createProject(manifestPath, ""), nil
	}

	version := ""
	for _, pg := range proj.PropertyGroups {
		if pg.TargetFramework != "" {
			version = extractDotNetVersion(pg.TargetFramework)
			break
		}
		if pg.TargetFrameworks != "" {
			// Multiple frameworks, take the first one
			frameworks := strings.Split(pg.TargetFrameworks, ";")
			if len(frameworks) > 0 {
				version = extractDotNetVersion(frameworks[0])
			}
			break
		}
	}

	return d.createProject(manifestPath, version), nil
}

func (d *dotNetDetector) detectSolutionFile(manifestPath string, content []byte) (*models.Project, error) {
	contentStr := string(content)

	// .sln files are text-based, check for Visual Studio signature
	if !strings.Contains(contentStr, "Microsoft Visual Studio") {
		return nil, nil
	}

	// Verify the solution references at least one .NET project file.
	// .sln files are shared between .NET and C++ (vcxproj) projects,
	// so we must check for actual .NET project references.
	hasDotNetProject := strings.Contains(contentStr, ".csproj") ||
		strings.Contains(contentStr, ".fsproj") ||
		strings.Contains(contentStr, ".vbproj")
	if !hasDotNetProject {
		return nil, nil
	}

	return d.createProject(manifestPath, ""), nil
}

func (d *dotNetDetector) createProject(manifestPath, version string) *models.Project {
	dir := filepath.Dir(manifestPath)
	name := filepath.Base(dir)

	// For .csproj files, use the file name as project name
	ext := strings.ToLower(filepath.Ext(manifestPath))
	if ext == ".csproj" || ext == ".fsproj" || ext == ".vbproj" {
		name = strings.TrimSuffix(filepath.Base(manifestPath), ext)
	}

	return &models.Project{
		Name:         name,
		Path:         dir,
		Runtime:      models.Runtime{Type: models.RuntimeDotNet, Version: version},
		ManifestFile: filepath.Base(manifestPath),
		SourcePaths:  []string{"."},
	}
}

// extractDotNetVersion extracts version from target framework moniker.
// Examples: net8.0 -> 8.0, net6.0-windows -> 6.0, netcoreapp3.1 -> 3.1
func extractDotNetVersion(tfm string) string {
	tfm = strings.TrimSpace(strings.ToLower(tfm))

	// Handle net5.0+ format
	re := regexp.MustCompile(`^net(\d+\.?\d*)`)
	if matches := re.FindStringSubmatch(tfm); len(matches) > 1 {
		return matches[1]
	}

	// Handle netcoreapp format
	re = regexp.MustCompile(`^netcoreapp(\d+\.?\d*)`)
	if matches := re.FindStringSubmatch(tfm); len(matches) > 1 {
		return matches[1]
	}

	// Handle netstandard format
	re = regexp.MustCompile(`^netstandard(\d+\.?\d*)`)
	if matches := re.FindStringSubmatch(tfm); len(matches) > 1 {
		return "standard " + matches[1]
	}

	return tfm
}
