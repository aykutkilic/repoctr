#!/bin/bash
set -e

# Build script for repo-ctr

BINARY="repo-ctr"
VERSION="${VERSION:-dev}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_step() {
    echo -e "${GREEN}==>${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}Warning:${NC} $1"
}

print_error() {
    echo -e "${RED}Error:${NC} $1"
}

usage() {
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  build       Build the binary (default)"
    echo "  test        Run tests"
    echo "  install     Install to \$GOPATH/bin"
    echo "  clean       Remove build artifacts"
    echo "  fmt         Format code"
    echo "  lint        Run linter"
    echo "  coverage    Generate coverage report"
    echo "  help        Show this help message"
}

cmd_build() {
    print_step "Building ${BINARY}..."
    go build -ldflags "-s -w -X main.Version=${VERSION}" -o "${BINARY}" ./cmd/repo-ctr
    echo -e "${GREEN}Built:${NC} ${BINARY}"
}

cmd_test() {
    print_step "Running tests..."
    go test -v ./...
}

cmd_install() {
    print_step "Installing ${BINARY}..."
    go install -ldflags "-s -w -X main.Version=${VERSION}" ./cmd/repo-ctr
    echo -e "${GREEN}Installed to:${NC} $(go env GOPATH)/bin/${BINARY}"
}

cmd_clean() {
    print_step "Cleaning build artifacts..."
    rm -f "${BINARY}"
    rm -f coverage.out coverage.html
    echo "Done."
}

cmd_fmt() {
    print_step "Formatting code..."
    go fmt ./...
    echo "Done."
}

cmd_lint() {
    print_step "Running linter..."
    if command -v golangci-lint &> /dev/null; then
        golangci-lint run
    else
        print_warning "golangci-lint not installed. Install with:"
        echo "  go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@latest"
        exit 1
    fi
}

cmd_coverage() {
    print_step "Generating coverage report..."
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    echo -e "${GREEN}Coverage report:${NC} coverage.html"
}

# Main
case "${1:-build}" in
    build)
        cmd_build
        ;;
    test)
        cmd_test
        ;;
    install)
        cmd_install
        ;;
    clean)
        cmd_clean
        ;;
    fmt)
        cmd_fmt
        ;;
    lint)
        cmd_lint
        ;;
    coverage)
        cmd_coverage
        ;;
    help|--help|-h)
        usage
        ;;
    *)
        print_error "Unknown command: $1"
        usage
        exit 1
        ;;
esac
