package stats

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"repoctr/internal/emoji"
	"repoctr/pkg/models"
)

// Reporter formats and outputs project statistics.
type Reporter struct {
	writer io.Writer
}

// NewReporter creates a new stats reporter.
func NewReporter(w io.Writer) *Reporter {
	return &Reporter{writer: w}
}

// Report outputs statistics for a list of project stats.
func (r *Reporter) Report(stats []*models.ProjectStats) {
	r.ReportWithOptions(stats, false)
}

// ReportWithOptions outputs statistics for a list of project stats with options.
func (r *Reporter) ReportWithOptions(stats []*models.ProjectStats, allFiles bool) {
	for _, s := range stats {
		r.reportProjectWithOptions(s, 0, allFiles)
	}

	// Print grand totals if multiple projects
	if len(stats) > 1 || (len(stats) == 1 && len(stats[0].Children) > 0) {
		totals := r.calculateTotals(stats)
		r.printSeparator()
		fmt.Fprintf(r.writer, "\n📊 GRAND TOTALS\n")
		r.printSeparator()
		fmt.Fprintf(r.writer, "   Files:      %d\n", totals.TotalFiles)
		fmt.Fprintf(r.writer, "   Folders:    %d\n", totals.TotalFolders)
		fmt.Fprintf(r.writer, "   Lines:      %d\n", totals.TotalLines)
		fmt.Fprintf(r.writer, "   Code:       %d\n", totals.CodeLines)
		fmt.Fprintf(r.writer, "   Blank:      %d\n", totals.BlankLines)
		fmt.Fprintf(r.writer, "   Size:       %s\n", formatSize(totals.TotalSize))
	}
}

func (r *Reporter) reportProjectWithOptions(stats *models.ProjectStats, depth int, allFiles bool) {
	indent := strings.Repeat("  ", depth)
	project := stats.Project

	// Project header
	r.printSeparator()
	techEmoji := emoji.Map(project.Runtime.Type)
	fmt.Fprintf(r.writer, "\n%s📁 %s %s (%s", indent, project.Name, techEmoji, project.Runtime.Type)
	if project.Runtime.Version != "" {
		fmt.Fprintf(r.writer, " %s", project.Runtime.Version)
	}
	fmt.Fprintf(r.writer, ")\n")
	fmt.Fprintf(r.writer, "%s   Path: %s\n", indent, project.Path)
	r.printSeparator()

	// Statistics table
	fmt.Fprintf(r.writer, "%s   %-12s %s\n", indent, "Files:", fmt.Sprintf("%d", stats.TotalFiles))
	fmt.Fprintf(r.writer, "%s   %-12s %s\n", indent, "Folders:", fmt.Sprintf("%d", stats.TotalFolders))
	fmt.Fprintf(r.writer, "%s   %-12s %s\n", indent, "Total Lines:", fmt.Sprintf("%d", stats.TotalLines))
	fmt.Fprintf(r.writer, "%s   %-12s %s\n", indent, "Code Lines:", fmt.Sprintf("%d", stats.CodeLines))
	fmt.Fprintf(r.writer, "%s   %-12s %s\n", indent, "Blank Lines:", fmt.Sprintf("%d", stats.BlankLines))
	fmt.Fprintf(r.writer, "%s   %-12s %s\n", indent, "Total Size:", formatSize(stats.TotalSize))

	// Files listing
	var filesToShow []models.FileStats
	var title string
	if allFiles && len(stats.AllFiles) > 0 {
		filesToShow = stats.AllFiles
		title = fmt.Sprintf("All %d files:", len(stats.AllFiles))
	} else if len(stats.LargestFiles) > 0 {
		filesToShow = stats.LargestFiles
		title = fmt.Sprintf("Top %d largest files:", len(stats.LargestFiles))
	}

	if len(filesToShow) > 0 {
		fmt.Fprintf(r.writer, "\n%s   %s\n", indent, title)
		for i, f := range filesToShow {
			relPath, _ := filepath.Rel(project.Path, f.Path)
			if relPath == "" {
				relPath = filepath.Base(f.Path)
			}
			fmt.Fprintf(r.writer, "%s     %d. %s (%d lines)\n", indent, i+1, relPath, f.Lines)
		}
	}

	// Report children
	for _, child := range stats.Children {
		fmt.Fprintln(r.writer)
		r.reportProjectWithOptions(child, depth+1, allFiles)
	}
}

func (r *Reporter) printSeparator() {
	fmt.Fprintf(r.writer, "%s\n", strings.Repeat("─", 60))
}

func (r *Reporter) calculateTotals(stats []*models.ProjectStats) *models.ProjectStats {
	totals := &models.ProjectStats{}

	var aggregate func([]*models.ProjectStats)
	aggregate = func(list []*models.ProjectStats) {
		for _, s := range list {
			totals.TotalFiles += s.TotalFiles
			totals.TotalFolders += s.TotalFolders
			totals.TotalLines += s.TotalLines
			totals.BlankLines += s.BlankLines
			totals.CodeLines += s.CodeLines
			totals.TotalSize += s.TotalSize
			aggregate(s.Children)
		}
	}

	aggregate(stats)
	return totals
}

// formatSize formats bytes into human-readable format.
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
