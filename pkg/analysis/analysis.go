package analysis

import (
	"math"
	"sort"
	"strings"

	"github.com/baptiste-b-pegasys/upgradebot/pkg/github"
)

// GetAnalysis - create analysis that will provide
// * all PRs merged in the new version (including risk assessment, files changed, packages changed, etc)
// * the list of all files changed (including risk assessment and linked PR where the file was changed)
func GetAnalysis(tagCompare github.TagCompare, filesChangedByQuorum []string, expectedFileConflicts []string) Analysis {
	analysis := Analysis{}
	analysis.PrStats = make([]PullRequestStats, len(tagCompare.PullRequests))

	// pre-processing
	mapFileAssessment := make(map[string]Assessment)
	for _, file := range filesChangedByQuorum {
		mapFileAssessment[file] = Warning
	}
	for _, file := range expectedFileConflicts {
		mapFileAssessment[file] = Conflict
	}

	// processing & ordering PRs
	for i, pr := range tagCompare.PullRequests {
		analysis.PrStats[i] = getPullRequestStats(pr, mapFileAssessment)
	}

	sort.SliceStable(analysis.PrStats, func(i, j int) bool {
		return analysis.PrStats[i].Data.ClosedAt < analysis.PrStats[j].Data.ClosedAt
	})

	analysis.FileStats = getChangedFilesStats(tagCompare, mapFileAssessment)

	return analysis
}

func getPullRequestStats(pr github.PullRequest, mapFileAssessment map[string]Assessment) PullRequestStats {
	stats := PullRequestStats{}

	stats.Data = pr.Data

	mapPackageChanged := make(map[string]int)

	for _, file := range pr.Files {
		stats.Assessment = Good
		if val, ok := mapFileAssessment[file.Filename]; stats.Assessment == Good && ok {
			if stats.Assessment != Conflict {
				stats.Assessment = val
			}
		}

		stats.LinesAddedCount += file.Additions
		stats.LinesRemovedCount += file.Deletions

		lastIndex := strings.LastIndex(file.Filename, "/")
		packagePath := file.Filename
		if lastIndex > 0 {
			packagePath = file.Filename[0:lastIndex]
		}
		mapPackageChanged[packagePath] = mapPackageChanged[packagePath] + 1

		switch file.Status {
		case "added":
			stats.FilesAddedCount += 1
		case "modified":
			stats.FilesModifiedCount += 1
		default:
			stats.FilesRemovedCount += 1
		}
	}

	sort.SliceStable(pr.Files, func(i, j int) bool {
		return pr.Files[i].GetTotalModifications() > pr.Files[j].GetTotalModifications()
	})

	stats.TopFilesChanged = pr.Files[0:int(math.Min(float64(len(pr.Files)), 5))]

	stats.TopPackagesChanged = make([]PackageStats, len(mapPackageChanged))

	i := 0
	for k, v := range mapPackageChanged {
		stats.TopPackagesChanged[i] = PackageStats{
			Name:  k,
			Count: v,
		}
		i++
	}

	sort.SliceStable(stats.TopPackagesChanged, func(i, j int) bool {
		return stats.TopPackagesChanged[i].Count > stats.TopPackagesChanged[j].Count
	})

	return stats
}

func getChangedFilesStats(tagCompare github.TagCompare, mapFileAssessment map[string]Assessment) []ChangedFileStats {
	prsPerFile := make(map[string][]github.PullRequestData)
	filePerFile := make(map[string]github.File)

	for _, file := range tagCompare.Files {
		prsPerFile[file.Filename] = make([]github.PullRequestData, 0)
		filePerFile[file.Filename] = file
	}

	for _, pr := range tagCompare.PullRequests {
		for _, file := range pr.Files {
			prsPerFile[file.Filename] = append(prsPerFile[file.Filename], pr.Data)
		}
	}

	stats := make([]ChangedFileStats, len(prsPerFile))

	i := 0
	for name, v := range prsPerFile {
		stats[i] = ChangedFileStats{AssociatedPRs: v, File: filePerFile[name], Assessment: mapFileAssessment[name]}
		i++
	}

	sort.SliceStable(stats, func(i, j int) bool {
		return stats[i].File.GetTotalModifications() > stats[j].File.GetTotalModifications()
	})

	return stats
}
