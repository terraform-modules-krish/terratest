// Package git allows to interact with Git.
package git

import (
	"os/exec"
	"strings"

	"github.com/terraform-modules-krish/terratest/modules/testing"
)

// GetCurrentBranchName retrieves the current branch name.
func GetCurrentBranchName(t testing.TestingT) string {
	out, err := GetCurrentBranchNameE(t)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// GetCurrentBranchNameE retrieves the current branch name.
func GetCurrentBranchNameE(t testing.TestingT) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	bytes, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bytes)), nil
}
