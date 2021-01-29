package main

import (
	"fmt"
	"strings"
	"time"
	upgradebot "upgradebot/src"
)

func main() {
	fmt.Println("Gather information from Go-Ethereum release to prepare a upstream upgrade")

	githubAPI := upgradebot.NewGithubAPI()

	upgradebot.CloneQuorumRepository()
	defer upgradebot.ClearQuorumRepository()

	baseTag := upgradebot.GetBaseGethTag()
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

	filesChangedByQuorum := upgradebot.GetChangedFilesAgainstGethBaseVersion(baseTag)
	expectedFileConflicts := upgradebot.GetConflictsFilesAgainstGethTargetVersion(targetTag)

	branchName := fmt.Sprintf("upgrade/go-ethereum/%s-%s", targetTag, time.Now().Format("2006102150405"))
	upgradebot.CreateBranchFromGethTag(targetTag, branchName)

	tagCompare := githubAPI.GetTagCompare(baseTag, targetTag)

	analysis := upgradebot.GetAnalysis(tagCompare, filesChangedByQuorum, expectedFileConflicts)

	builder := strings.Builder{}

	builder.WriteString(upgradebot.CreateMarkdownHeader())
	builder.WriteString("\n\n")
	builder.WriteString(upgradebot.CreateMarkdownReleaseSection(releaseData))
	builder.WriteString("\n\n")
	builder.WriteString(upgradebot.CreateMarkdownAnalysisSection(analysis))
	builder.WriteString("\n\n")

	githubAPI.CreatePullRequest(branchName, releaseData, builder.String())

	fmt.Println("Done")
}
