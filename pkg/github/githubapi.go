package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"
	"upgradebot/config"
)

const PullRequestTitleFormat = "[Upgrade] Go-Ethereum release %s"

type GithubAPI interface {
	GetGethReleaseData(tag string) ReleaseData
	GetGethTagComparison(base string, target string) TagCompare
	GetNextReleaseFrom(baseTag string) ReleaseData
	CreateQuorumPullRequest(branchName string, data ReleaseData, prBody string) PullRequestData
	FindOpenUpgradePullRequest(targetTag string) *PullRequestData
}

type apiImpl struct {
	httpAdapter *httpAdapter
}

func NewGithubAPI() GithubAPI {
	client := newHttpAdapter()
	return &apiImpl{
		httpAdapter: client,
	}
}

// GetNextReleaseFrom - get the next go-ethereum release after a specific version/tag
func (api *apiImpl) GetNextReleaseFrom(baseTag string) ReleaseData {
	releases := api.GetAllGethReleases()
	releaseIndex := 0

	for i, r := range releases {
		if r.Tag == baseTag {
			releaseIndex = i - 1
		}
	}

	if releaseIndex < 0 {
		log.Fatal("Next release error notn found")
	}

	return releases[releaseIndex]
}

// GetAllGethReleases - get all go-ethereum releases
func (api *apiImpl) GetAllGethReleases() []ReleaseData {
	body, _ := api.httpAdapter.sendGetRequest(config.GethGithubAPIUrl + "/releases")
	var data []ReleaseData

	jsonErr := json.Unmarshal(body, &data)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return data
}

// GetGethReleaseData - get go-ethereum release data based on a tag
func (api *apiImpl) GetGethReleaseData(tag string) ReleaseData {
	url := fmt.Sprintf("%s/releases/tags/%s", config.GethGithubAPIUrl, tag)

	body, _ := api.httpAdapter.sendGetRequest(url)
	data := ReleaseData{}

	jsonErr := json.Unmarshal(body, &data)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return data
}

// GetGethTagComparison - compare two geth tags and extract PR merged and files changed
func (api *apiImpl) GetGethTagComparison(base string, target string) TagCompare {
	commitChanges := api.getCommitChanges(base, target)
	prsData := api.getPullRequests(commitChanges)
	return TagCompare{PullRequests: prsData, Files: commitChanges.Files}
}

// CreateQuorumPullRequest - create PR in the quorum repo
func (api *apiImpl) CreateQuorumPullRequest(branchName string, data ReleaseData, prBody string) PullRequestData {
	title := fmt.Sprintf(PullRequestTitleFormat, data.Tag)
	createPrBody := CreatePullRequest{Title: title, Body: prBody, Base: "master", Head: branchName, Draft: true}
	jsonBody, _ := json.Marshal(createPrBody)
	response, _ := api.httpAdapter.sendPostRequest(config.QuorumAPIUrl+"/pulls", bytes.NewBuffer(jsonBody))

	result := PullRequestData{}
	jsonErr := json.Unmarshal(response, &result)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return result
}

func (api *apiImpl) FindOpenUpgradePullRequest(targetTag string) *PullRequestData {
	title := fmt.Sprintf(PullRequestTitleFormat, targetTag)

	response, _ := api.httpAdapter.sendGetRequest(config.QuorumAPIUrl + "/pulls?state=open&per_page=100")

	var result []PullRequestData
	jsonErr := json.Unmarshal(response, &result)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	for _, pr := range result {
		if pr.Title == title {
			return &pr
		}
	}
	return nil
}

func (api *apiImpl) getPullRequests(commitChanges CommitChanges) []PullRequest {
	prsData := api.getPullRequestDataFromCommits(commitChanges)

	pullRequests := make([]PullRequest, len(prsData))

	for i, prData := range prsData {
		files := api.getPullRequestFiles(prData)
		pr := PullRequest{Data: prData, Files: files}
		pullRequests[i] = pr
	}
	return pullRequests
}

func (api *apiImpl) getPullRequestFiles(prData PullRequestData) []File {
	url := fmt.Sprintf("%s/pulls/%d/files", config.GethGithubAPIUrl, prData.Number)

	body, _ := api.httpAdapter.sendGetRequest(url)
	var prFiles []File
	jsonErr := json.Unmarshal(body, &prFiles)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return prFiles
}

func (api *apiImpl) getPullRequestDataFromCommits(commitChanges CommitChanges) []PullRequestData {
	length := len(commitChanges.Commits)
	requestDataArray := make([]PullRequestData, 0)

	shas := make([]string, length)
	for i, c := range commitChanges.Commits {
		shas[i] = c.Sha[0:7]
	}

	for i := 0; i < length; i = i + 28 {
		end := math.Min(28, float64(length-i))
		requests := api.getPullRequestsData(shas[i : i+int(end)])
		requestDataArray = append(requestDataArray, requests...)
	}

	uniquePrs := make(map[int]bool)
	result := make([]PullRequestData, 0)

	// filter duplicates
	for _, pr := range requestDataArray {
		if uniquePrs[pr.Number] {
			continue
		}
		uniquePrs[pr.Number] = true
		result = append(result, pr)
	}

	return result
}

func (api *apiImpl) getPullRequestsData(shas []string) []PullRequestData {
	concatenatedSha := strings.Join(shas, "+")

	url := fmt.Sprintf("%s/search/issues?q=repo:ethereum/go-ethereum+is:pr+is:merged+merged+%s", config.GithubAPIUrl, concatenatedSha)
	body, _ := api.httpAdapter.sendGetRequest(url)

	prResult := struct {
		Items []PullRequestData
	}{}
	jsonErr := json.Unmarshal(body, &prResult)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return prResult.Items
}

func (api *apiImpl) getCommitChanges(base string, target string) CommitChanges {
	url := fmt.Sprintf("%s/compare/%s...%s", config.GethGithubAPIUrl, base, target)
	body, _ := api.httpAdapter.sendGetRequest(url)

	releaseCompare := CommitChanges{}
	jsonErr := json.Unmarshal(body, &releaseCompare)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return releaseCompare
}
