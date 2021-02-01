package git

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"upgradebot/config"
)

// CloneQuorumRepository - clone the repository of Quorum locally and add the go-ethereum remote as `geth`
func CloneQuorumRepository() {
	err := exec.Command("git", "clone", config.QuorumGitRepo, config.QuorumRepoFolder).Run()
	if err != nil {
		log.Fatal(err)
	}

	// load geth tags
	executeGitCommandOnRepo("remote", "add", "geth", config.GethGitRepo)
	executeGitCommandOnRepo("fetch", "geth", "--tags")
}

// ClearQuorumRepository - delete the repository folder
func ClearQuorumRepository() {
	exec.Command("rm", "-rf", config.QuorumRepoFolder).Run()
}

// CreateBranchFromGethTag - create a branch from a geth tag and push the branch to the remote quorum
func CreateBranchFromGethTag(targetTag string, branchName string) {
	executeGitCommandOnRepo("checkout", "tags/"+targetTag, "-b", branchName)
	executeGitCommandOnRepo("push", "-u", "origin", branchName)
}

/**
GetBaseGethTag - Get current version of go-ethereum merged into Quorum

params/version.go

VersionMajor = 1        // Major version component of the current release
VersionMinor = 9        // Minor version component of the current release
VersionPatch = 8        // Patch version component of the current release
VersionMeta  = "stable" // Version metadata to append to the version string

*/
func GetBaseGethTag() string {
	matcherMajor, _ := regexp.Compile(`VersionMajor = (\d+)`)
	matcherMinor, _ := regexp.Compile(`VersionMinor = (\d+)`)
	matcherPatch, _ := regexp.Compile(`VersionPatch = (\d+)`)
	out, err := ioutil.ReadFile(config.QuorumRepoFolder + config.QuorumVersionFilePath)
	if err != nil {
		log.Fatal("Error reading file "+config.QuorumVersionFilePath, err)
	}
	fileStr := string(out)

	if !matcherMajor.MatchString(fileStr) || !matcherMinor.MatchString(fileStr) || !matcherPatch.MatchString(fileStr) {
		log.Fatal("Failed to find the Geth version inside " + config.QuorumVersionFilePath)
	}

	majorVersion := matcherMajor.FindStringSubmatch(fileStr)[1]
	minorVersion := matcherMinor.FindStringSubmatch(fileStr)[1]
	patchVersion := matcherPatch.FindStringSubmatch(fileStr)[1]

	return fmt.Sprintf("v%s.%s.%s", majorVersion, minorVersion, patchVersion)
}

// GetConflictsFilesAgainstGethTargetVersion - Get the list of filenames that will have conflicts between Quorum master and the target geth tag
func GetConflictsFilesAgainstGethTargetVersion(targetGethTag string) []string {
	executeGitCommandOnRepo("merge", "--no-commit", "--no-ff", targetGethTag)
	defer executeGitCommandOnRepo("merge", "--abort")

	output, err := executeGitCommandOnRepo("diff", "--name-only", "--diff-filter=U")
	if err != nil {
		log.Fatal(err)
	}

	return strings.Split(string(output), "\n")
}

// GetChangedFilesAgainstGethBaseVersion - Get the list of filenames that were changed by quorum when comparing with the same geth tag currently merged into quorum
func GetChangedFilesAgainstGethBaseVersion(baseGethTag string) []string {
	output, _ := executeGitCommandOnRepo("diff", "--name-only", baseGethTag)
	return strings.Split(string(output), "\n")
}

func executeGitCommandOnRepo(arg ...string) ([]byte, error) {
	cmd := exec.Command("git", arg...)
	cmd.Dir = config.QuorumRepoFolder
	log.Println(cmd.String())
	return cmd.Output()
}
