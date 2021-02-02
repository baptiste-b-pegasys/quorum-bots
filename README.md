## Go-Ethereum upstream Upgrade tool for Quorum

Tool that creates automatically new Pull Requests for new Go-Ethereum versions into Quorum. Moreover, it analyses the PR and files changed, creating a table into the PR description with some assessment of the probability of issues or conflicts.

This tool is configured to run daily and target the https://github.com/ConsenSys/quorum repository, creating new PR when there are new versions of [Go-Ethereum](https://github.com/ethereum/go-ethereum).

### Run

You need to set environment variables to access the github API:
 * `GITHUB_USERNAME`: github username
 * `GITHUB_USER_TOKEN`: token created for that username with access to checkout projects and create PRs.

Run project:
`make run`

