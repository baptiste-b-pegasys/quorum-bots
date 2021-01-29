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

/**
params/version.go

VersionMajor = 1        // Major version component of the current release
VersionMinor = 9        // Minor version component of the current release
VersionPatch = 8        // Patch version component of the current release
VersionMeta  = "stable" // Version metadata to append to the version string

*/

func CloneQuorumRepository() {
	fmt.Println("Cloning quorum")
	exec.Command("git", "clone", config.QuorumGitRepo, config.QuorumRepoFolder).Run()

	// load geth tags
	fmt.Println("Adding go-ethereum remote")
	executeGitCommandOnRepo("remote", "add", "geth", config.GethGitRepo)
	fmt.Println("Getting geth tags")
	executeGitCommandOnRepo("fetch", "geth", "--tags")
}

func ClearQuorumRepository() {
	exec.Command("rm", "-rf", config.QuorumRepoFolder).Run()
}

func CreateBranchFromGethTag(targetTag string, branchName string) {
	executeGitCommandOnRepo("checkout", "tags/"+targetTag, "-b", branchName)
	executeGitCommandOnRepo("push", "-u", "origin", branchName)
}

func GetBaseGethTag() string {
	matcherMajor, _ := regexp.Compile("VersionMajor = (\\d+)")
	matcherMinor, _ := regexp.Compile("VersionMinor = (\\d+)")
	matcherPatch, _ := regexp.Compile("VersionPatch = (\\d+)")
	out, err := ioutil.ReadFile(config.QuorumRepoFolder + config.QuorumVersionFilePath)
	if err != nil {
		log.Fatal("Error reading file " + config.QuorumVersionFilePath)
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

func GetConflictsFilesAgainstGethTargetVersion(targetGethTag string) []string {
	executeGitCommandOnRepo("merge", "--no-commit", "--no-ff", targetGethTag)
	defer executeGitCommandOnRepo("merge", "--abort")

	output, _ := executeGitCommandOnRepo("diff", "--name-only", "--diff-filter=U")

	return strings.Split(string(output), "\n")
}

func GetChangedFilesAgainstGethBaseVersion(baseGethTag string) []string {
	output, _ := executeGitCommandOnRepo("diff", "--name-only", baseGethTag)
	return strings.Split(string(output), "\n")
}

func executeGitCommandOnRepo(arg ...string) ([]byte, error) {
	cmd := exec.Command("git", arg...)
	cmd.Dir = config.QuorumRepoFolder
	fmt.Println(cmd.String())
	return cmd.Output()
}
