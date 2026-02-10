package detector

import (
	"testing"

	"repoctr/pkg/models"
)

func TestGoDetector(t *testing.T) {
	d := NewGoDetector()

	tests := []struct {
		name        string
		content     string
		wantName    string
		wantVersion string
	}{
		{
			name: "basic go.mod",
			content: `module github.com/example/myapp

go 1.21

require (
	github.com/spf13/cobra v1.8.0
)`,
			wantName:    "myapp",
			wantVersion: "1.21",
		},
		{
			name: "simple module",
			content: `module myproject

go 1.22`,
			wantName:    "myproject",
			wantVersion: "1.22",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project, err := d.Detect("go.mod", []byte(tt.content))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if project == nil {
				t.Fatal("expected project, got nil")
			}
			if project.Name != tt.wantName {
				t.Errorf("name = %q, want %q", project.Name, tt.wantName)
			}
			if project.Runtime.Version != tt.wantVersion {
				t.Errorf("version = %q, want %q", project.Runtime.Version, tt.wantVersion)
			}
			if project.Runtime.Type != models.RuntimeGo {
				t.Errorf("type = %q, want %q", project.Runtime.Type, models.RuntimeGo)
			}
		})
	}
}

func TestPythonDetector_Pyproject(t *testing.T) {
	d := NewPythonDetector()

	content := `[project]
name = "mypackage"
requires-python = ">=3.9"

[build-system]
requires = ["setuptools>=61.0"]
`

	project, err := d.Detect("pyproject.toml", []byte(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project == nil {
		t.Fatal("expected project, got nil")
	}
	if project.Name != "mypackage" {
		t.Errorf("name = %q, want %q", project.Name, "mypackage")
	}
	if project.Runtime.Version != "3.9+" {
		t.Errorf("version = %q, want %q", project.Runtime.Version, "3.9+")
	}
}

func TestPythonDetector_Poetry(t *testing.T) {
	d := NewPythonDetector()

	content := `[tool.poetry]
name = "poetry-project"
python = "^3.10"
`

	project, err := d.Detect("pyproject.toml", []byte(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project == nil {
		t.Fatal("expected project, got nil")
	}
	if project.Name != "poetry-project" {
		t.Errorf("name = %q, want %q", project.Name, "poetry-project")
	}
	if project.Runtime.Version != "3.10+" {
		t.Errorf("version = %q, want %q", project.Runtime.Version, "3.10+")
	}
}

func TestJavaScriptDetector(t *testing.T) {
	d := NewJavaScriptDetector()

	content := `{
  "name": "my-app",
  "version": "1.0.0",
  "engines": {
    "node": ">=18.0.0"
  }
}`

	project, err := d.Detect("package.json", []byte(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project == nil {
		t.Fatal("expected project, got nil")
	}
	if project.Name != "my-app" {
		t.Errorf("name = %q, want %q", project.Name, "my-app")
	}
	if project.Runtime.Version != ">=18.0.0" {
		t.Errorf("version = %q, want %q", project.Runtime.Version, ">=18.0.0")
	}
	if project.Runtime.Type != models.RuntimeJavaScript {
		t.Errorf("type = %q, want %q", project.Runtime.Type, models.RuntimeJavaScript)
	}
}

func TestRustDetector(t *testing.T) {
	d := NewRustDetector()

	content := `[package]
name = "my-crate"
version = "0.1.0"
edition = "2021"
rust-version = "1.70"
`

	project, err := d.Detect("Cargo.toml", []byte(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project == nil {
		t.Fatal("expected project, got nil")
	}
	if project.Name != "my-crate" {
		t.Errorf("name = %q, want %q", project.Name, "my-crate")
	}
	if project.Runtime.Version != "1.70" {
		t.Errorf("version = %q, want %q", project.Runtime.Version, "1.70")
	}
}

func TestDotNetDetector(t *testing.T) {
	d := NewDotNetDetector()

	content := `<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <TargetFramework>net8.0</TargetFramework>
  </PropertyGroup>
</Project>`

	project, err := d.Detect("MyApp.csproj", []byte(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project == nil {
		t.Fatal("expected project, got nil")
	}
	if project.Name != "MyApp" {
		t.Errorf("name = %q, want %q", project.Name, "MyApp")
	}
	if project.Runtime.Version != "8.0" {
		t.Errorf("version = %q, want %q", project.Runtime.Version, "8.0")
	}
}

func TestJavaDetector_Maven(t *testing.T) {
	d := NewJavaDetector()

	content := `<?xml version="1.0"?>
<project>
  <artifactId>my-java-app</artifactId>
  <properties>
    <java.version>17</java.version>
  </properties>
</project>`

	project, err := d.Detect("pom.xml", []byte(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project == nil {
		t.Fatal("expected project, got nil")
	}
	if project.Name != "my-java-app" {
		t.Errorf("name = %q, want %q", project.Name, "my-java-app")
	}
	if project.Runtime.Version != "17" {
		t.Errorf("version = %q, want %q", project.Runtime.Version, "17")
	}
}

func TestDotNetDetector_SlnWithCsproj(t *testing.T) {
	d := NewDotNetDetector()

	content := `Microsoft Visual Studio Solution File, Format Version 12.00
# Visual Studio Version 17
Project("{FAE04EC0-301F-11D3-BF4B-00C04F79EFBC}") = "MyApp", "MyApp\MyApp.csproj", "{GUID}"
EndProject`

	project, err := d.Detect("dir/MyApp.sln", []byte(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project == nil {
		t.Fatal("expected project, got nil")
	}
	if project.Runtime.Type != models.RuntimeDotNet {
		t.Errorf("type = %q, want %q", project.Runtime.Type, models.RuntimeDotNet)
	}
}

func TestDotNetDetector_SlnWithVcxprojOnly(t *testing.T) {
	d := NewDotNetDetector()

	// A .sln that only references .vcxproj (C++ project) should NOT be detected as .NET
	content := `Microsoft Visual Studio Solution File, Format Version 12.00
# Visual Studio Version 17
Project("{8BC9CEB8-8B4A-11D0-8D11-00A0C91BC942}") = "MyCppApp", "MyCppApp\MyCppApp.vcxproj", "{GUID}"
EndProject`

	project, err := d.Detect("dir/MyCppApp.sln", []byte(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project != nil {
		t.Errorf("expected nil for C++ .sln, got project with runtime %q", project.Runtime.Type)
	}
}

func TestCppDetector_Vcxproj(t *testing.T) {
	d := NewCppDetector()

	content := `<?xml version="1.0" encoding="utf-8"?>
<Project DefaultTargets="Build" xmlns="http://schemas.microsoft.com/developer/msbuild/2003">
  <PropertyGroup Label="Globals">
    <RootNamespace>SingeltonTest</RootNamespace>
  </PropertyGroup>
  <Import Project="$(VCTargetsPath)\Microsoft.Cpp.Default.props" />
  <ItemGroup>
    <ClCompile Include="main.cpp" />
  </ItemGroup>
</Project>`

	project, err := d.Detect("dir/SingeltonTest.vcxproj", []byte(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project == nil {
		t.Fatal("expected project, got nil")
	}
	if project.Name != "SingeltonTest" {
		t.Errorf("name = %q, want %q", project.Name, "SingeltonTest")
	}
	if project.Runtime.Type != models.RuntimeCpp {
		t.Errorf("type = %q, want %q", project.Runtime.Type, models.RuntimeCpp)
	}
}

func TestRegistry_CppSlnNotDotNet(t *testing.T) {
	r := NewRegistry()

	// When a .sln only references .vcxproj files, the registry should NOT detect it as .NET
	content := `Microsoft Visual Studio Solution File, Format Version 12.00
# Visual Studio Version 17
Project("{8BC9CEB8-8B4A-11D0-8D11-00A0C91BC942}") = "MyCpp", "MyCpp\MyCpp.vcxproj", "{GUID}"
EndProject`

	project, err := r.DetectProject("dir/MyCpp.sln", []byte(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should be nil since .sln is not a manifest for C++ detector (it uses .vcxproj directly)
	if project != nil {
		t.Errorf("expected nil for C++ .sln via registry, got runtime %q", project.Runtime.Type)
	}
}

func TestRegistry(t *testing.T) {
	r := NewRegistry()

	patterns := r.GetManifestPatterns()
	if len(patterns) == 0 {
		t.Error("expected manifest patterns, got none")
	}

	// Check that common manifest files are included
	expected := []string{"go.mod", "package.json", "Cargo.toml", "pyproject.toml", "pom.xml", "pubspec.yaml"}
	for _, exp := range expected {
		found := false
		for _, p := range patterns {
			if p == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected pattern %q not found in registry", exp)
		}
	}
}
