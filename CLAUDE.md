# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go backend service (structure to be defined as project develops).

## Build & Development Commands

```bash
# Initialize module (first time setup)
go mod init repoctr

# Build
go build ./...

# Run tests
go test ./...

# Run a single test
go test -run TestName ./path/to/package

# Run with verbose output
go test -v ./...

# Lint (if using golangci-lint)
golangci-lint run

# Format code
go fmt ./...
```

## Code Style

- Ensure no errors or warnings in written code
- Follow standard Go conventions (gofmt, effective Go guidelines)
