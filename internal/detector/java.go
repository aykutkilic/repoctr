package detector

import (
	"encoding/xml"
	"path/filepath"
	"regexp"
	"strings"

	"repoctr/pkg/models"
)

type javaDetector struct{}

func NewJavaDetector() Detector {
	return &javaDetector{}
}

func (d *javaDetector) Name() string {
	return "Java"
}

func (d *javaDetector) RuntimeType() models.RuntimeType {
	return models.RuntimeJava
}

func (d *javaDetector) ManifestFiles() []string {
	return []string{"pom.xml", "build.gradle", "build.gradle.kts"}
}

func (d *javaDetector) Detect(manifestPath string, content []byte) (*models.Project, error) {
	filename := filepath.Base(manifestPath)

	switch filename {
	case "pom.xml":
		return d.detectPomXml(manifestPath, content)
	case "build.gradle", "build.gradle.kts":
		return d.detectGradle(manifestPath, content)
	}

	return nil, nil
}

// pomXml represents the structure of a pom.xml file.
type pomXml struct {
	XMLName    xml.Name `xml:"project"`
	ArtifactID string   `xml:"artifactId"`
	Name       string   `xml:"name"`
	Properties struct {
		JavaVersion         string `xml:"java.version"`
		MavenCompilerSource string `xml:"maven.compiler.source"`
	} `xml:"properties"`
}

func (d *javaDetector) detectPomXml(manifestPath string, content []byte) (*models.Project, error) {
	// Check if this looks like a Maven POM
	if !strings.Contains(string(content), "<project") {
		return nil, nil
	}

	var pom pomXml
	if err := xml.Unmarshal(content, &pom); err != nil {
		// If XML parsing fails, still detect as Java project
		return d.createProject(manifestPath, "", ""), nil
	}

	name := pom.Name
	if name == "" {
		name = pom.ArtifactID
	}

	version := pom.Properties.JavaVersion
	if version == "" {
		version = pom.Properties.MavenCompilerSource
	}

	return d.createProject(manifestPath, name, version), nil
}

func (d *javaDetector) detectGradle(manifestPath string, content []byte) (*models.Project, error) {
	contentStr := string(content)

	// Check for common Gradle patterns
	if !strings.Contains(contentStr, "plugins") && !strings.Contains(contentStr, "apply plugin") &&
		!strings.Contains(contentStr, "dependencies") {
		return nil, nil
	}

	// Try to extract sourceCompatibility or java version
	version := ""

	// Try sourceCompatibility = '11' or sourceCompatibility = "11"
	compatRe := regexp.MustCompile(`sourceCompatibility\s*=\s*['"]?(\d+)['"]?`)
	if matches := compatRe.FindStringSubmatch(contentStr); len(matches) > 1 {
		version = matches[1]
	}

	// Try JavaVersion.VERSION_11
	javaVersionRe := regexp.MustCompile(`JavaVersion\.VERSION_(\d+)`)
	if matches := javaVersionRe.FindStringSubmatch(contentStr); len(matches) > 1 {
		version = matches[1]
	}

	// Try java { toolchain { languageVersion = JavaLanguageVersion.of(17) } }
	toolchainRe := regexp.MustCompile(`JavaLanguageVersion\.of\((\d+)\)`)
	if matches := toolchainRe.FindStringSubmatch(contentStr); len(matches) > 1 {
		version = matches[1]
	}

	return d.createProject(manifestPath, "", version), nil
}

func (d *javaDetector) createProject(manifestPath, name, version string) *models.Project {
	dir := filepath.Dir(manifestPath)
	if name == "" {
		name = filepath.Base(dir)
	}

	return &models.Project{
		Name:           name,
		Path:           dir,
		Runtime:        models.Runtime{Type: models.RuntimeJava, Version: version},
		ManifestFile:   filepath.Base(manifestPath),
		SourcePaths:    []string{"src/main/java", "src"},
		SrcIgnorePaths: []string{"target", "build"},
	}
}
