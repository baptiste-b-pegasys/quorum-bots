package config

import (
	"os"
	"sync"
)

var once sync.Once

type Config struct {
	GithubAPIUrl string

	QuorumGitRepo string
	QuorumAPIUrl  string

	GethGitRepo      string
	GethGithubAPIUrl string

	GithubUsername  string
	GithubUserToken string

	QuorumRepoFolder      string
	QuorumVersionFilePath string
}

var (
	instance *Config
)

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{
			GithubAPIUrl: "https://api.github.com",

			GithubUsername:  os.Getenv("GITHUB_USERNAME"),
			GithubUserToken: os.Getenv("GITHUB_USER_TOKEN"),

			GethGitRepo:      "https://github.com/ethereum/go-ethereum.git",
			GethGithubAPIUrl: "https://api.github.com/repos/ethereum/go-ethereum",

			QuorumGitRepo: "git@github.com:ConsenSys/quorum.git",
			QuorumAPIUrl:  "https://api.github.com/repos/ConsenSys/quorum",

			// For experimentation with the private Quorum repository
			//QuorumGitRepo: "git@github.com:ConsenSysQuorum/quorum.git",
			//QuorumAPIUrl:  "https://api.github.com/repos/ConsenSysQuorum/quorum",

			QuorumRepoFolder:      "tmp-quorum-repo",
			QuorumVersionFilePath: "/params/version.go",
		}

	})
	return instance
}
