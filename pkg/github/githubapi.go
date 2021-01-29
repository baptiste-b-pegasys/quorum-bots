package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strings"
	"upgradebot/config"
)

const USERNAME = "ricardolyn"
const TOKEN = "8b3d3c6b486590135699987e7e760de92575c8bf"

const PullRequestTitleFormat = "[Upgrade] Go-Ethereum release %s"

type apiImpl struct {
	httpClient *http.Client
}

type GithubAPI interface {
	GetReleaseData(tag string) ReleaseData
	GetTagCompare(base string, target string) TagCompare
	GetNextReleaseFrom(baseTag string) ReleaseData
	CreatePullRequest(branchName string, data ReleaseData, prBody string) PullRequestData
	FindOpenUpgradePullRequest(targetTag string) *PullRequestData
}

func NewGithubAPI() GithubAPI {
	client := &http.Client{}
	return &apiImpl{
		httpClient: client,
	}
}

func (api *apiImpl) GetNextReleaseFrom(baseTag string) ReleaseData {
	releases := api.GetAllReleases()
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

func (api *apiImpl) GetAllReleases() []ReleaseData {
	body, _ := api.sendGetRequest(config.GethGithubAPIUrl + "/releases")
	var data []ReleaseData

	jsonErr := json.Unmarshal(body, &data)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return data
}

func (api *apiImpl) GetReleaseData(tag string) ReleaseData {
	url := fmt.Sprintf("%s/releases/tags/%s", config.GethGithubAPIUrl, tag)

	body, _ := api.sendGetRequest(url)
	data := ReleaseData{}

	jsonErr := json.Unmarshal(body, &data)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return data
}

func (api *apiImpl) GetTagCompare(base string, target string) TagCompare {
	commitChanges := api.getCommitChanges(base, target)
	prsData := api.getPullRequests(commitChanges)
	return TagCompare{PullRequests: prsData, Files: commitChanges.Files}
}

func (api *apiImpl) CreatePullRequest(branchName string, data ReleaseData, prBody string) PullRequestData {
	title := fmt.Sprintf(PullRequestTitleFormat, data.Tag)
	createPrBody := CreatePullRequest{Title: title, Body: prBody, Base: "master", Head: branchName, Draft: true}
	jsonBody, _ := json.Marshal(createPrBody)
	response, _ := api.sendPostRequest(config.QuorumAPIUrl+"/pulls", bytes.NewBuffer(jsonBody))

	result := PullRequestData{}
	jsonErr := json.Unmarshal(response, &result)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return result
}

func (api *apiImpl) FindOpenUpgradePullRequest(targetTag string) *PullRequestData {
	title := fmt.Sprintf(PullRequestTitleFormat, targetTag)

	response, _ := api.sendGetRequest(config.QuorumAPIUrl + "/pulls?state=open&per_page=100")

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

	body, _ := api.sendGetRequest(url)
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

	url := fmt.Sprintf("https://api.com/search/issues?q=repo:ethereum/go-ethereum+is:pr+is:merged+merged+%s", concatenatedSha)
	body, _ := api.sendGetRequest(url)

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
	body, _ := api.sendGetRequest(url)

	releaseCompare := CommitChanges{}
	jsonErr := json.Unmarshal(body, &releaseCompare)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return releaseCompare
}

func (api *apiImpl) sendPostRequest(url string, body io.Reader) ([]byte, error) {
	fmt.Println(url)
	req, _ := http.NewRequest("POST", url, body)
	return api.sendRequest(req)
}

func (api *apiImpl) sendGetRequest(url string) ([]byte, error) {
	fmt.Println(url)
	req, _ := http.NewRequest("GET", url, nil)
	return api.sendRequest(req)
}

func (api *apiImpl) sendRequest(req *http.Request) ([]byte, error) {
	req.SetBasicAuth(USERNAME, TOKEN)
	req.Header.Add("Accept", "application/vnd.v3+json")
	resp, _ := api.httpClient.Do(req)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// print HTTP headers
	//for name, values := range resp.Header {
	//	// Loop over all values for the name.
	//	for _, value := range values {
	//		fmt.Println(name, value)
	//	}
	//}

	return body, nil
}
