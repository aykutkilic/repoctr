package detector

import (
	"path/filepath"
	"regexp"
	"strings"

	"repoctr/pkg/models"
)

type cppDetector struct{}

func NewCppDetector() Detector {
	return &cppDetector{}
}

func (d *cppDetector) Name() string {
	return "C/C++"
}

func (d *cppDetector) RuntimeType() models.RuntimeType {
	return models.RuntimeCpp
}

func (d *cppDetector) ManifestFiles() []string {
	return []string{"CMakeLists.txt", "Makefile", "meson.build", "*.vcxproj"}
}

func (d *cppDetector) Detect(manifestPath string, content []byte) (*models.Project, error) {
	filename := filepath.Base(manifestPath)

	ext := strings.ToLower(filepath.Ext(manifestPath))

	switch {
	case filename == "CMakeLists.txt":
		return d.detectCMake(manifestPath, content)
	case filename == "Makefile":
		return d.detectMakefile(manifestPath, content)
	case filename == "meson.build":
		return d.detectMeson(manifestPath, content)
	case ext == ".vcxproj":
		return d.detectVcxproj(manifestPath, content)
	}

	return nil, nil
}

func (d *cppDetector) detectCMake(manifestPath string, content []byte) (*models.Project, error) {
	contentStr := string(content)

	// Check for cmake_minimum_required or project() to confirm it's a CMake file
	if !strings.Contains(contentStr, "cmake_minimum_required") && !strings.Contains(contentStr, "project(") {
		return nil, nil
	}

	// Try to extract project name
	name := ""
	projectRe := regexp.MustCompile(`project\s*\(\s*(\w+)`)
	if matches := projectRe.FindStringSubmatch(contentStr); len(matches) > 1 {
		name = matches[1]
	}

	// Try to extract C++ standard
	version := ""
	stdRe := regexp.MustCompile(`CMAKE_CXX_STANDARD\s+(\d+)`)
	if matches := stdRe.FindStringSubmatch(contentStr); len(matches) > 1 {
		version = "C++" + matches[1]
	}

	// Also check for C standard
	if version == "" {
		cStdRe := regexp.MustCompile(`CMAKE_C_STANDARD\s+(\d+)`)
		if matches := cStdRe.FindStringSubmatch(contentStr); len(matches) > 1 {
			version = "C" + matches[1]
		}
	}

	return d.createProject(manifestPath, name, version), nil
}

func (d *cppDetector) detectMakefile(manifestPath string, content []byte) (*models.Project, error) {
	contentStr := string(content)

	// Check for C/C++ specific indicators in Makefile
	// Must have C/C++ compiler variable references ($(CC), $(CXX)) or direct compiler calls
	isCppMakefile := false

	// Check for C/C++ compiler variable references - strong indicator
	if strings.Contains(contentStr, "$(CC)") || strings.Contains(contentStr, "$(CXX)") {
		isCppMakefile = true
	}

	// Check for direct compiler calls with compile flags
	compilerRe := regexp.MustCompile(`\b(gcc|g\+\+|clang|clang\+\+)\s+.*-[co]\b`)
	if compilerRe.MatchString(contentStr) {
		isCppMakefile = true
	}

	// Check for C/C++ standard flags - definitive indicator
	if strings.Contains(contentStr, "-std=c") {
		isCppMakefile = true
	}

	// Check for common C/C++ build variables with assignments (more specific)
	cBuildVarsRe := regexp.MustCompile(`\b(CFLAGS|CXXFLAGS|CPPFLAGS)\s*[:+]?=`)
	if cBuildVarsRe.MatchString(contentStr) {
		isCppMakefile = true
	}

	if !isCppMakefile {
		return nil, nil
	}

	// Try to extract C++ standard from flags
	version := ""
	stdRe := regexp.MustCompile(`-std=c\+\+(\d+)`)
	if matches := stdRe.FindStringSubmatch(contentStr); len(matches) > 1 {
		version = "C++" + matches[1]
	}

	// Also check for C standard
	if version == "" {
		cStdRe := regexp.MustCompile(`-std=c(\d+)`)
		if matches := cStdRe.FindStringSubmatch(contentStr); len(matches) > 1 {
			version = "C" + matches[1]
		}
	}

	return d.createProject(manifestPath, "", version), nil
}

func (d *cppDetector) detectMeson(manifestPath string, content []byte) (*models.Project, error) {
	contentStr := string(content)

	// Check for meson project() call
	if !strings.Contains(contentStr, "project(") {
		return nil, nil
	}

	// Try to extract project name
	name := ""
	projectRe := regexp.MustCompile(`project\s*\(\s*'([^']+)'`)
	if matches := projectRe.FindStringSubmatch(contentStr); len(matches) > 1 {
		name = matches[1]
	}

	return d.createProject(manifestPath, name, ""), nil
}

func (d *cppDetector) detectVcxproj(manifestPath string, content []byte) (*models.Project, error) {
	contentStr := string(content)

	// .vcxproj files are MSBuild XML for C/C++ projects.
	// Check for C++ specific indicators.
	if !strings.Contains(contentStr, "<ClCompile") && !strings.Contains(contentStr, "Microsoft.Cpp") {
		return nil, nil
	}

	// Try to extract project name from RootNamespace or filename
	name := ""
	nameRe := regexp.MustCompile(`<RootNamespace>([^<]+)</RootNamespace>`)
	if matches := nameRe.FindStringSubmatch(contentStr); len(matches) > 1 {
		name = matches[1]
	}

	return d.createProject(manifestPath, name, ""), nil
}

func (d *cppDetector) createProject(manifestPath, name, version string) *models.Project {
	dir := filepath.Dir(manifestPath)
	if name == "" {
		name = filepath.Base(dir)
	}

	return &models.Project{
		Name:           name,
		Path:           dir,
		Runtime:        models.Runtime{Type: models.RuntimeCpp, Version: version},
		ManifestFile:   filepath.Base(manifestPath),
		SourcePaths:    []string{"src", "include", "."},
		SrcIgnorePaths: []string{"build", "cmake-build-debug", "cmake-build-release"},
	}
}
