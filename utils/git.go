
package utils

import (
	"os/exec"
	"strings"
)

// GetGitUserName attempts to get the user's name from the git config.
func GetGitUserName() string {
	cmd := exec.Command("git", "config", "--get", "user.name")
	output, err := cmd.Output()
	if err != nil {
		return "" // Return empty string if git is not installed or name is not set
	}
	return strings.TrimSpace(string(output))
}
