package main

import (
	"fmt"
	"strings"
	"time"
	"upgradebot/pkg/analysis"
	"upgradebot/pkg/git"
	"upgradebot/pkg/github"
	"upgradebot/pkg/markdown"
)

func main() {
	fmt.Println("Gather information from Go-Ethereum release to prepare a upstream upgrade")

	githubAPI := github.NewGithubAPI()

	git.CloneQuorumRepository()
	defer git.ClearQuorumRepository()

	baseTag := git.GetBaseGethTag()
	releaseData := githubAPI.GetNextReleaseFrom(baseTag)
	targetTag := releaseData.Tag

	openPr := githubAPI.FindOpenUpgradePullRequest(targetTag)

	if openPr != nil {
		fmt.Printf("There is already a PR open with the name \"%s\" and number %d. Ignore\n", openPr.Title, openPr.Number)
		return
	}

	if baseTag == targetTag {
		fmt.Printf("We are already in the latest version %s. Ignore\n", baseTag)
		return
	}

	fmt.Printf("Base version: %s. Target Version: %s\n", baseTag, targetTag)

	filesChangedByQuorum := git.GetChangedFilesAgainstGethBaseVersion(baseTag)
	expectedFileConflicts := git.GetConflictsFilesAgainstGethTargetVersion(targetTag)

	branchName := fmt.Sprintf("upgrade/go-ethereum/%s-%s", targetTag, time.Now().Format("2006102150405"))
	git.CreateBranchFromGethTag(targetTag, branchName)

	tagCompare := githubAPI.GetGethTagComparison(baseTag, targetTag)

	analysis := analysis.GetAnalysis(tagCompare, filesChangedByQuorum, expectedFileConflicts)

	builder := strings.Builder{}

	builder.WriteString(markdown.CreateMarkdownHeader())
	builder.WriteString("\n\n")
	builder.WriteString(markdown.CreateMarkdownReleaseSection(releaseData))
	builder.WriteString("\n\n")
	builder.WriteString(markdown.CreateMarkdownAnalysisSection(analysis))
	builder.WriteString("\n\n")

	githubAPI.CreateQuorumPullRequest(branchName, releaseData, builder.String())

	fmt.Println("Done")
}
