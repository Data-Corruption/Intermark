package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Data-Corruption/blog"
)

func GitInstalled() bool {
	cmd := exec.Command("git", "--version")
	if err := cmd.Run(); err != nil {
		blog.Errorf("Git is not installed: %v", err)
		return false
	} else {
		blog.Info("Git is installed")
		return true
	}
}

// GitFileDiff checks if the given file has changed since the given commit.
// It returns true if the file has changed, false if not, and an error if any.
// If commitHash is empty, it returns true.
func GitFileDiff(repoPath, filePath, commitHash string) (bool, error) {
	if commitHash == "" {
		return true, nil
	}

	// Validate input parameters and check if the paths exist
	if repoPath == "" || filePath == "" {
		return false, fmt.Errorf("repository path and file path cannot be empty")
	}
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return false, fmt.Errorf("repository path does not exist: %w", err)
	}
	if _, err := os.Stat(filepath.Join(repoPath, filePath)); os.IsNotExist(err) {
		return false, fmt.Errorf("file does not exist: %w", err)
	}

	// Prepare and run the git diff command, capturing output
	cmd := exec.Command("git", "diff", "--exit-code", commitHash, "--", filePath)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If the error is an exit error and the exit code is 1, the file has changed. Otherwise, return the error
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				return true, nil
			}
		}
		return false, fmt.Errorf("error running git diff: %w\nOutput: %s", err, strings.TrimSpace(string(output)))
	}
	return false, nil
}

func GitCommitHash(repoPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error getting commit hash: %w\nOutput: %s", err, strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output)), nil
}

func GitClone(repoURL, repoPath string) (string, error) {
	if err := os.MkdirAll(repoPath, os.ModePerm); err != nil {
		return "", fmt.Errorf("error creating repository path: %w", err)
	}
	// modify the repoURL to use the configured SSH host if it contains 'github.com'
	repoURL = strings.Replace(repoURL, "github.com", Config.ContentRepo.SshHost, 1)
	// clone the repo
	cmd := exec.Command("git", "clone", repoURL, ".")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running git clone: %w\nOutput: %s", err, strings.TrimSpace(string(output)))
	}
	blog.Debugf("Clone output: %s", strings.TrimSpace(string(output)))
	return GitCommitHash(repoPath)
}

// GitReset resets the repository to the latest commit on the branch set in the config and returns the commit hash.
func GitReset(repoPath string) (string, error) {
	if err := os.MkdirAll(repoPath, os.ModePerm); err != nil {
		return "", fmt.Errorf("error creating repository path: %w", err)
	}
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
		return "", fmt.Errorf("repository path does not contain a .git directory")
	}

	cmd := exec.Command("git", "fetch", "origin", Config.ContentRepo.Branch)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running git fetch origin %s: %w\nOutput: %s", Config.ContentRepo.Branch, err, strings.TrimSpace(string(output)))
	}
	blog.Debugf("Fetch output: %s", strings.TrimSpace(string(output)))

	cmd = exec.Command("git", "reset", "--hard", "origin/"+Config.ContentRepo.Branch)
	cmd.Dir = repoPath
	output, err = cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running git reset --hard origin/%s: %w\nOutput: %s", Config.ContentRepo.Branch, err, strings.TrimSpace(string(output)))
	}
	blog.Debugf("Reset output: %s", strings.TrimSpace(string(output)))
	return GitCommitHash(repoPath)
}
