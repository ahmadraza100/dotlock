package model

import (
	"fmt"
)

// RenderShell renders a map to shell compatible KEY="value" lines.
func RenderShell(m map[string][]byte) string {
	out := ""
	for k, v := range m {
		out += fmt.Sprintf("%s=\"%s\"\n", k, string(v))
	}
	return out
}

// RenderDocker renders KEY=value lines for a docker .env file
func RenderDocker(m map[string][]byte) string {
	out := ""
	for k, v := range m {
		out += fmt.Sprintf("%s=%s\n", k, string(v))
	}
	return out
}

// RenderGitHubActions renders succinct commands for setting secrets (user can pipe)
func RenderGitHubActions(m map[string][]byte) string {
	out := ""
	for k, v := range m {
		out += fmt.Sprintf("gh secret set %s --body \"%s\"\n", k, string(v))
	}
	return out
}
