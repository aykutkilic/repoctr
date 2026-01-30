# repo-ctr

A Go CLI tool for discovering projects and calculating lines of code (LOC) statistics in repositories.

## Features

- **Auto-detection** of projects based on manifest files (go.mod, package.json, Cargo.toml, etc.)
- **Version extraction** from project configuration files
- **Hierarchical project tree** for monorepos with nested projects
- **LOC statistics** including total lines, code lines, blank lines, and file sizes
- **Top 5 largest files** per project
- **gitignore-aware** traversal with sensible defaults
- **Machine-readable output** in YAML, JSON, XML, or CSV formats

## Supported Runtimes

| Runtime | Manifest Files | Version Source |
|---------|---------------|----------------|
| Go | `go.mod` | `go 1.xx` directive |
| Python | `pyproject.toml`, `setup.py`, `requirements.txt` | `requires-python` or poetry config |
| JavaScript | `package.json` | `engines.node` |
| TypeScript | `package.json` + `tsconfig.json` | `engines.node` |
| Java | `pom.xml`, `build.gradle`, `build.gradle.kts` | `java.version` or `sourceCompatibility` |
| .NET | `*.csproj`, `*.sln`, `*.fsproj`, `*.vbproj` | `<TargetFramework>` XML element |
| Rust | `Cargo.toml` | `rust-version` or `edition` |
| Dart | `pubspec.yaml` | `environment.sdk` |
| C/C++ | `CMakeLists.txt`, `Makefile` | `CMAKE_CXX_STANDARD` or `-std=` flags |

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/repoctr.git
cd repoctr

# Build (outputs to current directory)
./build.sh build     # Linux/macOS
build.bat build      # Windows

# Run
./repo-ctr --help    # Linux/macOS
repo-ctr --help      # Windows

# Install to $GOPATH/bin for global access
./build.sh install   # Linux/macOS
build.bat install    # Windows
```

### Using go install

```bash
go install github.com/yourusername/repoctr/cmd/repo-ctr@latest
```

### Prerequisites

- Go 1.21 or later

## Usage

### Quick Start

```bash
repo-ctr identify .   # Discover projects
repo-ctr              # Show stats (if projects.yaml exists)
```

### Initialize a Configuration

Create a `projects.yaml` template:

```bash
repo-ctr init
```

### Discover Projects

Scan directories to auto-discover projects:

```bash
# Scan current directory
repo-ctr identify .

# Scan multiple directories
repo-ctr identify ./src ./lib ./packages

# Custom output file
repo-ctr identify . -o my-projects.yaml
```

### View Statistics

Calculate and display LOC statistics:

```bash
# Default: shows stats if projects.yaml exists
repo-ctr

# Explicit stats command
repo-ctr stats

# Using custom file
repo-ctr stats -f my-projects.yaml
```

### Machine-Readable Output

Export statistics in various formats for scripting and automation:

```bash
# YAML format (default for --machine)
repo-ctr stats --machine
repo-ctr stats --yaml

# JSON format
repo-ctr stats --json

# XML format
repo-ctr stats --xml

# CSV format (flat, no hierarchy)
repo-ctr stats --csv
```

Example JSON output:
```json
{
  "projects": [
    {
      "name": "my-app",
      "path": ".",
      "runtime": "Go",
      "version": "1.21",
      "files": 45,
      "folders": 12,
      "total_lines": 3500,
      "code_lines": 2800,
      "blank_lines": 700,
      "size_bytes": 125000
    }
  ],
  "totals": {
    "files": 45,
    "folders": 12,
    "total_lines": 3500,
    "code_lines": 2800,
    "blank_lines": 700,
    "size_bytes": 125000
  }
}
```

## Example Output

### Identify Command

```
Scanning /path/to/monorepo...
  Found 5 project(s)

Wrote 5 project(s) to /path/to/monorepo/projects.yaml
  - monorepo (Go 1.21)
    - api (Go 1.21)
    - web (TypeScript)
      - ui-components (TypeScript)
    - scripts (Python 3.11+)
```

### Stats Command

```
────────────────────────────────────────────────────────────

📁 api (Go 1.21)
   Path: services/api
────────────────────────────────────────────────────────────
   Files:       45
   Folders:     12
   Total Lines: 3500
   Code Lines:  2800
   Blank Lines: 700
   Total Size:  125.3 KB

   Top 5 largest files:
     1. handler.go (450 lines)
     2. service.go (380 lines)
     3. repository.go (320 lines)
     4. models.go (280 lines)
     5. middleware.go (220 lines)

────────────────────────────────────────────────────────────

📊 GRAND TOTALS
────────────────────────────────────────────────────────────
   Files:      156
   Folders:    42
   Lines:      12500
   Code:       10200
   Blank:      2300
   Size:       450.2 KB
```

## Configuration

The `projects.yaml` file defines the project structure:

```yaml
projects:
  - name: my-project
    path: .
    runtime:
      type: Go
      version: "1.21"
    manifest-file: go.mod
    source-paths:
      - .
    src-ignore-paths:
      - vendor
    children:
      - name: subproject
        path: packages/subproject
        runtime:
          type: TypeScript
        manifest-file: package.json
        source-paths:
          - src
        src-ignore-paths:
          - node_modules
          - dist
```

### Fields

| Field | Description |
|-------|-------------|
| `name` | Project display name |
| `path` | Relative path from repository root |
| `runtime.type` | Runtime type (Go, Python, TypeScript, etc.) |
| `runtime.version` | Runtime version (optional) |
| `manifest-file` | The manifest file that defines the project |
| `source-paths` | Directories to include in LOC counting |
| `src-ignore-paths` | Directories to exclude from LOC counting |
| `children` | Nested child projects |

## Default Ignored Paths

The following directories are always ignored during discovery and statistics:

- Version control: `.git`, `.svn`, `.hg`
- Dependencies: `node_modules`, `vendor`, `__pycache__`, `venv`, `.venv`
- Build outputs: `target`, `build`, `dist`, `bin`, `obj`
- IDE: `.idea`, `.vscode`, `.vs`
- OS files: `.DS_Store`, `Thumbs.db`

## Development

Using the build scripts:

```bash
# Linux/macOS
./build.sh build      # Build binary
./build.sh test       # Run tests
./build.sh fmt        # Format code
./build.sh lint       # Run linter
./build.sh coverage   # Generate coverage report
./build.sh clean      # Clean build artifacts
./build.sh help       # Show all commands

# Windows
build.bat build
build.bat test
build.bat fmt
build.bat lint
build.bat coverage
build.bat clean
build.bat help
```

Or use Go commands directly:

```bash
go build ./...                           # Build all packages
go test ./...                            # Run all tests
go test -v ./...                         # Run tests with verbose output
go test -coverprofile=coverage.out ./... # Generate coverage
go fmt ./...                             # Format code
golangci-lint run                        # Run linter
```

## Project Structure

```
repo-ctr/
├── cmd/repo-ctr/
│   ├── main.go           # Entry point
│   └── root.go           # Root command + subcommands
├── internal/
│   ├── cli/              # Command implementations
│   ├── detector/         # Runtime detectors
│   ├── discovery/        # Filesystem walker + hierarchy builder
│   ├── stats/            # LOC counter + reporter
│   └── ignore/           # Ignore pattern matcher
├── pkg/models/           # Shared types
├── build.sh              # Build script (Linux/macOS)
├── build.bat             # Build script (Windows)
├── go.mod
└── repo-ctr              # Built binary (after ./build.sh build)
```

## License

MIT License
