package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"repoctr/internal/cli"
)

const projectsFileName = "projects.yaml"

var rootCmd = &cobra.Command{
	Use:   "repo-ctr",
	Short: "Repository project discovery and LOC statistics tool",
	Long: `repo-ctr is a CLI tool for discovering projects in repositories
and calculating lines of code statistics.

It automatically detects various project types including:
  - Go (go.mod)
  - Python (pyproject.toml, setup.py, requirements.txt)
  - JavaScript/TypeScript (package.json)
  - Java (pom.xml, build.gradle)
  - .NET (*.csproj, *.sln)
  - Rust (Cargo.toml)
  - Dart (pubspec.yaml)
  - C/C++ (CMakeLists.txt, Makefile)

Usage:
  1. repo-ctr init              - Create a projects.yaml template
  2. repo-ctr identify .        - Auto-discover projects
  3. repo-ctr stats             - Show LOC statistics

If projects.yaml exists, running 'repo-ctr' without arguments shows stats.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If projects.yaml exists, run stats by default
		if _, err := os.Stat(projectsFileName); err == nil {
			return cli.RunStats(projectsFileName, false, "")
		}

		// Auto-discover projects and show stats
		fmt.Println("No projects.yaml found. Auto-discovering projects...")
		if err := cli.RunIdentify([]string{"."}, projectsFileName); err != nil {
			return err
		}

		// Check if any projects were discovered (projects.yaml should exist now)
		if _, err := os.Stat(projectsFileName); err != nil {
			fmt.Println("\nNo projects discovered. Use 'repo-ctr identify <path>' to scan a specific directory.")
			return nil
		}

		fmt.Println()
		return cli.RunStats(projectsFileName, false, "")
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(cli.NewInitCmd())
	rootCmd.AddCommand(cli.NewIdentifyCmd())
	rootCmd.AddCommand(cli.NewStatsCmd())
	rootCmd.AddCommand(cli.NewVersionCmd())
	rootCmd.AddCommand(cli.NewUpdateCmd())
}
