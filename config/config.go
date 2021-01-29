package config

const (
	GithubAPIUrl = "https://api.github.com"

	// Public
	//QuorumGitRepo = "https://github.com/ConsenSys/quorum.git"
	//QuorumAPIUrl     = "https://api.github.com/repos/ConsenSys/quorum"

	// Private - TODO: swap to public URLs when ready to be used and stable
	QuorumGitRepo = "git@github.com:ConsenSysQuorum/quorum.git"
	QuorumAPIUrl  = "https://api.github.com/repos/ConsenSysQuorum/quorum"

	// Go-Ethereum
	GethGitRepo      = "https://github.com/ethereum/go-ethereum.git"
	GethGithubAPIUrl = "https://api.github.com/repos/ethereum/go-ethereum"

	// Config
	QuorumRepoFolder      = "tmp-quorum-repo"
	QuorumVersionFilePath = "/params/version.go"
)
