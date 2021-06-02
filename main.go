package main

import (
	"fmt"
	"log"
	"strings"
	"time"
	"upgradebot/config"
	"upgradebot/pkg/analysis"
	"upgradebot/pkg/git"
	"upgradebot/pkg/github/http"
	"upgradebot/pkg/markdown"
)

func main() {
	log.Println("Gather information from Go-Ethereum release to prepare an upstream upgrade")

	cfg := config.GetConfig()
	githubAPI := http.NewGithub(cfg)
	git := git.NewGit(cfg)

	git.CloneQuorumRepository()
	defer git.ClearQuorumRepository()

	baseTag := git.GetBaseGethTag()
	releaseData := githubAPI.GetNextReleaseFrom(baseTag)
	targetTag := releaseData.Tag

	// Validate if we are already in the latest go-ethereum version
	if baseTag == targetTag {
		log.Printf("We are already in the latest version %s. Ignore\n", baseTag)
		return
	}

	// Validate if we don't have any PR already opened for an upgrade of the new version
	openPr := githubAPI.FindOpenUpgradePullRequest(targetTag)
	if openPr != nil {
		log.Printf("There is already a PR on %s. Ignore\n", openPr.HtmlUrl)
		return
	}

	log.Printf("Preparing release PR. Base version: %s. Target Version: %s\n", baseTag, targetTag)

	// Analyse the quorum and go-ethereum changes to provide an overview of new features and PRs
	filesChangedByQuorum := git.GetChangedFilesAgainstGethBaseVersion(baseTag)
	expectedFileConflicts := git.GetConflictsFilesAgainstGethTargetVersion(targetTag)
	tagCompare := githubAPI.GetGethTagComparison(baseTag, targetTag)
	analysis := analysis.GetAnalysis(tagCompare, filesChangedByQuorum, expectedFileConflicts)

	// Create PR body
	builder := strings.Builder{}
	builder.WriteString(markdown.CreateMarkdownHeader())
	builder.WriteString("\n\n")
	builder.WriteString(markdown.CreateMarkdownReleaseSection(releaseData))
	builder.WriteString("\n\n")
	builder.WriteString(markdown.CreateMarkdownAnalysisSection(analysis))
	builder.WriteString("\n\n")

	// Create new branch and the  upgrade PR
	branchName := fmt.Sprintf("upgrade/go-ethereum/%s-%s", targetTag, time.Now().Format("2006102150405"))
	git.CreateBranchFromGethTag(targetTag, branchName)
	createdPr := githubAPI.CreateQuorumPullRequest(branchName, releaseData, builder.String())

	log.Println("Done, PR: " + createdPr.HtmlUrl)
}
