package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"repoctr/internal/config"
	"repoctr/pkg/models"
)

// NewConfigCmd creates the config command group.
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage repository configuration",
		Long:  "Manage .repoctrconfig.yaml settings for exclusions and project overrides.",
	}

	cmd.AddCommand(
		newConfigInitCmd(),
		newConfigAddExcludeCmd(),
		newConfigShowCmd(),
	)

	return cmd
}

// newConfigInitCmd creates the 'config init' subcommand.
func newConfigInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a .repoctrconfig.yaml template",
		Long:  "Creates a .repoctrconfig.yaml template file in the current directory.",
		RunE:  runConfigInit,
	}

	return cmd
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	rootDir, _ := filepath.Abs(".")

	// Check if config already exists
	configPath := config.ConfigPath(rootDir)
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("%s already exists. Use 'repo-ctr config show' to view it", filepath.Base(configPath))
	}

	// Create template config
	templateConfig := &models.RepoCtrConfig{
		GlobalExcludes: []string{
			"# Examples of global exclusions applied to all projects:",
			"# **/*.test.js",
			"# **/__mocks__/**",
			"# **/generated/**",
		},
		ProjectOverrides: map[string]models.ProjectOverride{
			"lib": {
				ExcludePatterns: []string{
					"# examples/**",
					"# test_data/**",
				},
			},
		},
	}

	// Save config
	if err := config.SaveConfig(rootDir, templateConfig); err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	absPath := config.ConfigPath(rootDir)
	absPath, _ = filepath.Abs(absPath)
	fmt.Printf("Created %s\n", absPath)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Edit .repoctrconfig.yaml to add your exclusion patterns")
	fmt.Println("  2. Run 'repo-ctr stats' to apply the exclusions")
	fmt.Println("  3. Use 'repo-ctr config show' to view current configuration")

	return nil
}

// newConfigAddExcludeCmd creates the 'config add-exclude' subcommand.
func newConfigAddExcludeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-exclude <pattern>",
		Short: "Add a global exclusion pattern",
		Long: `Add a gitignore-style pattern to global exclusions.
The pattern will be applied to all projects.

Examples:
  repo-ctr config add-exclude "**/*.test.js"
  repo-ctr config add-exclude "node_modules/**"
  repo-ctr config add-exclude "__pycache__"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigAddExclude(args[0])
		},
	}

	return cmd
}

func runConfigAddExclude(pattern string) error {
	rootDir, _ := filepath.Abs(".")

	// Load existing config
	cfg, err := config.LoadConfig(rootDir)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Add pattern
	cfg.GlobalExcludes = append(cfg.GlobalExcludes, pattern)

	// Save config
	if err := config.SaveConfig(rootDir, cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Added pattern: %s\n", pattern)
	fmt.Println("Run 'repo-ctr stats' to apply the new exclusion")

	return nil
}

// newConfigShowCmd creates the 'config show' subcommand.
func newConfigShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Display current configuration",
		Long:  "Shows the contents of .repoctrconfig.yaml.",
		RunE:  runConfigShow,
	}

	return cmd
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	rootDir, _ := filepath.Abs(".")

	configPath := config.ConfigPath(rootDir)
	configPath, _ = filepath.Abs(configPath)

	fmt.Printf("Configuration: %s\n\n", configPath)

	// Check if config exists
	if _, err := os.Stat(configPath); err != nil {
		fmt.Println("No .repoctrconfig.yaml found. Run 'repo-ctr config init' to create one.")
		return nil
	}

	// Display config
	data, _ := os.ReadFile(configPath)
	fmt.Print(string(data))

	return nil
}
