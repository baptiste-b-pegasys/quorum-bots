package git

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"regexp"
	"strings"

	"upgradebot/config"
)

type Git struct {
	config *config.Config
}

func NewGit(config *config.Config) *Git {
	return &Git{config: config}
}

// CloneQuorumRepository - clone the repository of Quorum locally and add the go-ethereum remote as `geth`
func (s *Git) CloneQuorumRepository() {
	fmt.Println("git version")
	cmd := exec.Command("git", "version")
	out, err := cmd.Output()
	if checkCmdError("git version", cmd, out, err) {
		return
	} else {
		fmt.Println(string(out))
	}

	fmt.Println("git clone " + s.config.QuorumGitRepo + " " + s.config.QuorumRepoFolder)
	cmd = exec.Command("git", "clone", s.config.QuorumGitRepo, s.config.QuorumRepoFolder)
	out, err = cmd.Output()
	if checkCmdError("git clone", cmd, out, err) {
		return
	} else {
		fmt.Println(string(out))
	}

	// add quorum bot fork
	out, err = s.executeGitCommandOnRepo("remote", "add", "quorumbot", s.config.QuorumBotGitRepo)
	if checkCmdError("git clone", cmd, out, err) {
		return
	} else {
		fmt.Println(string(out))
	}

	// load geth tags
	cmd = s.buildGitCommandOnRepo("remote", "add", "geth", s.config.GethGitRepo)
	out, err = cmd.Output()
	if checkCmdError("git remote add", cmd, out, err) {
		return
	} else {
		fmt.Println(string(out))
	}

	cmd = s.buildGitCommandOnRepo("fetch", "geth", "--tags", "-f")
	out, err = cmd.Output()
	if checkCmdError("git fetch tags", cmd, out, err) {
		return
	} else {
		fmt.Println(string(out))
	}
}

func checkCmdError(reason string, cmd *exec.Cmd, out []byte, err error) bool {
	if err == nil {
		return false
	}
	log.Printf("env: %+v", cmd.Env)
	switch err := err.(type) {
	case *exec.ExitError:
		log.Fatalf("%s: %s: %v\n%s\n%s", reason, cmd.String(), err, string(out), string(err.Stderr))
	default:
		log.Fatalf("%s: %s: %v\n%s", reason, cmd.String(), err, string(out))
	}
	return true
}

// ClearQuorumRepository - delete the repository folder
func (s *Git) ClearQuorumRepository() {
	exec.Command("rm", "-rf", s.config.QuorumRepoFolder).Run()
}

// CreateBranchFromGethTag - create a branch from a geth tag and push the branch to the remote quorum
func (s *Git) CreateBranchFromGethTag(targetTag string, branchName string) {
	s.executeGitCommandOnRepo("checkout", "tags/"+targetTag, "-b", branchName)
	s.executeGitCommandOnRepo("push", "-u", "quorumbot", branchName)
}

/**
GetBaseGethTag - Get current version of go-ethereum merged into Quorum

params/version.go

VersionMajor = 1        // Major version component of the current release
VersionMinor = 9        // Minor version component of the current release
VersionPatch = 8        // Patch version component of the current release
VersionMeta  = "stable" // Version metadata to append to the version string

*/
func (s *Git) GetBaseGethTag() string {
	matcherMajor, _ := regexp.Compile(`VersionMajor = (\d+)`)
	matcherMinor, _ := regexp.Compile(`VersionMinor = (\d+)`)
	matcherPatch, _ := regexp.Compile(`VersionPatch = (\d+)`)
	out, err := ioutil.ReadFile(s.config.QuorumRepoFolder + s.config.QuorumVersionFilePath)
	if err != nil {
		log.Fatal("Error reading file "+s.config.QuorumVersionFilePath, err)
	}
	fileStr := string(out)

	if !matcherMajor.MatchString(fileStr) || !matcherMinor.MatchString(fileStr) || !matcherPatch.MatchString(fileStr) {
		log.Fatal("Failed to find the Geth version inside " + s.config.QuorumVersionFilePath)
	}

	majorVersion := matcherMajor.FindStringSubmatch(fileStr)[1]
	minorVersion := matcherMinor.FindStringSubmatch(fileStr)[1]
	patchVersion := matcherPatch.FindStringSubmatch(fileStr)[1]

	return fmt.Sprintf("v%s.%s.%s", majorVersion, minorVersion, patchVersion)
}

// GetConflictsFilesAgainstGethTargetVersion - Get the list of filenames that will have conflicts between Quorum master and the target geth tag
func (s *Git) GetConflictsFilesAgainstGethTargetVersion(targetGethTag string) []string {
	output, err := s.executeGitCommandOnRepo("merge", "--no-commit", "--no-ff", targetGethTag)
	if err != nil {
		log.Fatal(err)
		if exiterr, ok := err.(*exec.ExitError); ok {
			fmt.Println(string(exiterr.Stderr))
		}
	} else {
		fmt.Println(string(output))
	}
	defer s.executeGitCommandOnRepo("merge", "--abort")

	output, err = s.executeGitCommandOnRepo("diff", "--diff-filter=U", "--name-only")
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(string(output))
	}

	return strings.Split(string(output), "\n")
}

// GetChangedFilesAgainstGethBaseVersion - Get the list of filenames that were changed by quorum when comparing with the same geth tag currently merged into quorum
func (s *Git) GetChangedFilesAgainstGethBaseVersion(baseGethTag string) []string {
	output, err := s.executeGitCommandOnRepo("diff", "--name-only", baseGethTag)
	if err != nil {
		fmt.Println("error", err.Error())
	} else {
		fmt.Println(string(output))
	}
	return strings.Split(string(output), "\n")
}

func (s *Git) executeGitCommandOnRepo(arg ...string) ([]byte, error) {
	cmd := s.buildGitCommandOnRepo(arg...)
	log.Println(cmd.String())
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(cmd.Stderr.(*bytes.Buffer).String())
	}
	return out, err
}

func (s *Git) buildGitCommandOnRepo(arg ...string) *exec.Cmd {
	fmt.Println("git ", arg, " on dir ", s.config.QuorumRepoFolder)
	cmd := exec.Command("git", arg...)
	buf := make([]byte, 1024)
	w := bytes.NewBuffer(buf)
	cmd.Stderr = w
	cmd.Dir = s.config.QuorumRepoFolder
	return cmd
}
