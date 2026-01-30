package cli

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"repoctr/internal/stats"
	"repoctr/pkg/models"
)

// OutputFormat represents the machine-readable output format.
type OutputFormat string

const (
	FormatYAML OutputFormat = "yaml"
	FormatJSON OutputFormat = "json"
	FormatXML  OutputFormat = "xml"
	FormatCSV  OutputFormat = "csv"
)

// NewStatsCmd creates the stats command.
func NewStatsCmd() *cobra.Command {
	var inputFile string
	var machine bool
	var yamlOut, jsonOut, xmlOut, csvOut bool

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show LOC statistics for discovered projects",
		Long: `Reads projects.yaml and calculates lines of code statistics.
Shows total files, folders, lines, code lines, blank lines, and file sizes.
Also displays the top 5 largest files per project.

Use --machine to output in machine-readable format (default: yaml).
Supported formats: --yaml, --json, --xml, --csv`,
		RunE: func(cmd *cobra.Command, args []string) error {
			format := ""
			if yamlOut {
				format = "yaml"
			} else if jsonOut {
				format = "json"
			} else if xmlOut {
				format = "xml"
			} else if csvOut {
				format = "csv"
			}
			return RunStats(inputFile, machine, format)
		},
	}

	cmd.Flags().StringVarP(&inputFile, "file", "f", projectsFileName, "Projects configuration file")
	cmd.Flags().BoolVarP(&machine, "machine", "m", false, "Output in machine-readable format (default: yaml)")
	cmd.Flags().BoolVar(&yamlOut, "yaml", false, "Output in YAML format")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&xmlOut, "xml", false, "Output in XML format")
	cmd.Flags().BoolVar(&csvOut, "csv", false, "Output in CSV format")

	return cmd
}

// RunStats executes the stats command logic (exported for use by root command).
func RunStats(inputFile string, machine bool, format string) error {
	// Read projects.yaml
	data, err := os.ReadFile(inputFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s not found. Run 'repo-ctr init' or 'repo-ctr identify .' first", inputFile)
		}
		return fmt.Errorf("failed to read %s: %w", inputFile, err)
	}

	var config models.ProjectsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse %s: %w", inputFile, err)
	}

	if len(config.Projects) == 0 {
		fmt.Println("No projects found in", inputFile)
		return nil
	}

	// Get the directory containing projects.yaml as root
	rootDir, err := filepath.Abs(filepath.Dir(inputFile))
	if err != nil {
		rootDir = "."
	}

	// Create counter
	counter, err := stats.NewCounter(rootDir)
	if err != nil {
		return fmt.Errorf("failed to create stats counter: %w", err)
	}

	// Calculate stats for all projects
	projectStats, err := counter.CountHierarchy(config.Projects)
	if err != nil {
		return fmt.Errorf("failed to calculate statistics: %w", err)
	}

	// Determine output format
	outputFormat := determineFormat(machine, format)

	if outputFormat != "" {
		return outputMachineReadable(projectStats, outputFormat)
	}

	// Human-readable output
	reporter := stats.NewReporter(os.Stdout)
	reporter.Report(projectStats)

	return nil
}

func determineFormat(machine bool, format string) OutputFormat {
	// Check explicit format flags
	switch format {
	case "yaml":
		return FormatYAML
	case "json":
		return FormatJSON
	case "xml":
		return FormatXML
	case "csv":
		return FormatCSV
	}

	// If --machine flag is set without format, default to YAML
	if machine {
		return FormatYAML
	}

	return ""
}

// StatsOutput represents the machine-readable stats output.
type StatsOutput struct {
	XMLName  xml.Name             `xml:"statistics" json:"-" yaml:"-"`
	Projects []ProjectStatsOutput `yaml:"projects" json:"projects" xml:"project"`
	Totals   TotalsOutput         `yaml:"totals" json:"totals" xml:"totals"`
}

// ProjectStatsOutput represents stats for a single project.
type ProjectStatsOutput struct {
	Name         string               `yaml:"name" json:"name" xml:"name"`
	Path         string               `yaml:"path" json:"path" xml:"path"`
	Runtime      string               `yaml:"runtime" json:"runtime" xml:"runtime"`
	Version      string               `yaml:"version,omitempty" json:"version,omitempty" xml:"version,omitempty"`
	Files        int                  `yaml:"files" json:"files" xml:"files"`
	Folders      int                  `yaml:"folders" json:"folders" xml:"folders"`
	TotalLines   int                  `yaml:"total_lines" json:"total_lines" xml:"total_lines"`
	CodeLines    int                  `yaml:"code_lines" json:"code_lines" xml:"code_lines"`
	BlankLines   int                  `yaml:"blank_lines" json:"blank_lines" xml:"blank_lines"`
	SizeBytes    int64                `yaml:"size_bytes" json:"size_bytes" xml:"size_bytes"`
	LargestFiles []FileStatsOutput    `yaml:"largest_files,omitempty" json:"largest_files,omitempty" xml:"largest_file,omitempty"`
	Children     []ProjectStatsOutput `yaml:"children,omitempty" json:"children,omitempty" xml:"child,omitempty"`
}

// FileStatsOutput represents stats for a single file.
type FileStatsOutput struct {
	Path  string `yaml:"path" json:"path" xml:"path"`
	Lines int    `yaml:"lines" json:"lines" xml:"lines"`
}

// TotalsOutput represents the grand totals.
type TotalsOutput struct {
	Files      int   `yaml:"files" json:"files" xml:"files"`
	Folders    int   `yaml:"folders" json:"folders" xml:"folders"`
	TotalLines int   `yaml:"total_lines" json:"total_lines" xml:"total_lines"`
	CodeLines  int   `yaml:"code_lines" json:"code_lines" xml:"code_lines"`
	BlankLines int   `yaml:"blank_lines" json:"blank_lines" xml:"blank_lines"`
	SizeBytes  int64 `yaml:"size_bytes" json:"size_bytes" xml:"size_bytes"`
}

func outputMachineReadable(projectStats []*models.ProjectStats, format OutputFormat) error {
	output := buildStatsOutput(projectStats)

	switch format {
	case FormatYAML:
		return outputYAML(output)
	case FormatJSON:
		return outputJSON(output)
	case FormatXML:
		return outputXML(output)
	case FormatCSV:
		return outputCSV(projectStats)
	}

	return fmt.Errorf("unknown format: %s", format)
}

func buildStatsOutput(projectStats []*models.ProjectStats) StatsOutput {
	output := StatsOutput{
		Projects: convertProjectStats(projectStats),
		Totals:   calculateTotals(projectStats),
	}
	return output
}

func convertProjectStats(stats []*models.ProjectStats) []ProjectStatsOutput {
	var result []ProjectStatsOutput

	for _, s := range stats {
		p := ProjectStatsOutput{
			Name:       s.Project.Name,
			Path:       s.Project.Path,
			Runtime:    string(s.Project.Runtime.Type),
			Version:    s.Project.Runtime.Version,
			Files:      s.TotalFiles,
			Folders:    s.TotalFolders,
			TotalLines: s.TotalLines,
			CodeLines:  s.CodeLines,
			BlankLines: s.BlankLines,
			SizeBytes:  s.TotalSize,
		}

		for _, f := range s.LargestFiles {
			p.LargestFiles = append(p.LargestFiles, FileStatsOutput{
				Path:  filepath.Base(f.Path),
				Lines: f.Lines,
			})
		}

		if len(s.Children) > 0 {
			p.Children = convertProjectStats(s.Children)
		}

		result = append(result, p)
	}

	return result
}

func calculateTotals(stats []*models.ProjectStats) TotalsOutput {
	totals := TotalsOutput{}

	var aggregate func([]*models.ProjectStats)
	aggregate = func(list []*models.ProjectStats) {
		for _, s := range list {
			totals.Files += s.TotalFiles
			totals.Folders += s.TotalFolders
			totals.TotalLines += s.TotalLines
			totals.CodeLines += s.CodeLines
			totals.BlankLines += s.BlankLines
			totals.SizeBytes += s.TotalSize
			aggregate(s.Children)
		}
	}

	aggregate(stats)
	return totals
}

func outputYAML(output StatsOutput) error {
	encoder := yaml.NewEncoder(os.Stdout)
	encoder.SetIndent(2)
	return encoder.Encode(output)
}

func outputJSON(output StatsOutput) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputXML(output StatsOutput) error {
	fmt.Println(`<?xml version="1.0" encoding="UTF-8"?>`)
	encoder := xml.NewEncoder(os.Stdout)
	encoder.Indent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return err
	}
	fmt.Println()
	return nil
}

func outputCSV(projectStats []*models.ProjectStats) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	header := []string{"name", "path", "runtime", "version", "files", "folders", "total_lines", "code_lines", "blank_lines", "size_bytes"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Flatten and write all projects
	var writeProject func(*models.ProjectStats)
	writeProject = func(s *models.ProjectStats) {
		row := []string{
			s.Project.Name,
			s.Project.Path,
			string(s.Project.Runtime.Type),
			s.Project.Runtime.Version,
			strconv.Itoa(s.TotalFiles),
			strconv.Itoa(s.TotalFolders),
			strconv.Itoa(s.TotalLines),
			strconv.Itoa(s.CodeLines),
			strconv.Itoa(s.BlankLines),
			strconv.FormatInt(s.TotalSize, 10),
		}
		writer.Write(row)

		for _, child := range s.Children {
			writeProject(child)
		}
	}

	for _, s := range projectStats {
		writeProject(s)
	}

	return nil
}
