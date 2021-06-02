package github

type Github interface {
	GetGethReleaseData(tag string) ReleaseData
	GetGethTagComparison(base string, target string) TagCompare
	GetNextReleaseFrom(baseTag string) ReleaseData
	CreateQuorumPullRequest(branchName string, data ReleaseData, prBody string) (*PullRequestData, error)
	FindOpenUpgradePullRequest(targetTag string) *PullRequestData
	AddLabelsToIssue(issueNumber int, labels ...string) *LabelsRequestData
}
