package analysis

import (
	"upgradebot/pkg/github"
)

type PullRequestStats struct {
	Data github.PullRequestData

	FilesAddedCount    int
	FilesRemovedCount  int
	FilesModifiedCount int

	LinesAddedCount   int
	LinesRemovedCount int

	TopFilesChanged    []github.File
	TopPackagesChanged []PackageStats

	Assessment Assessment
}

type Assessment string

const (
	Good     Assessment = "Good"
	Warning             = "Warning"
	Conflict            = "Conflict"
)

type PackageStats struct {
	Name  string
	Count int
}

type ChangedFileStats struct {
	AssociatedPRs []github.PullRequestData
	File          github.File
	Assessment    Assessment
}

type Analysis struct {
	PrStats   []PullRequestStats
	FileStats []ChangedFileStats
}
