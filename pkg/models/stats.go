package models

// FileStats holds statistics for a single file.
type FileStats struct {
	Path       string
	Lines      int
	BlankLines int
	CodeLines  int
	Size       int64
}

// ProjectStats holds aggregated statistics for a project.
type ProjectStats struct {
	Project      *Project
	TotalFiles   int
	TotalFolders int
	TotalLines   int
	BlankLines   int
	CodeLines    int
	TotalSize    int64
	LargestFiles []FileStats
	AllFiles     []FileStats
	Children     []*ProjectStats
}
