// Package safe provides helpers to detect dangerous shell commands.
package safe

import "strings"

var dangerousPatterns = []string{
	"rm -rf /",
	"rm -rf /*",
	"> /dev/sda",
	"mkfs.",
	"dd if=",
	":(){ :|:& };:",
	"chmod -R 777 /",
	"chown -R",
	"> /etc/",
	"curl | sh",
	"wget -O - | sh",
}

var warnings = []string{
	"rm ",
	"rm -rf",
	"git push --force",
	"git reset --hard",
	"DROP ",
	"DELETE FROM",
	"TRUNCATE ",
	"shutdown",
	"reboot",
}

type Check struct {
	IsDangerous bool
	IsWarning   bool
	Pattern     string
}

func Audit(cmd string) Check {
	for _, p := range dangerousPatterns {
		if strings.Contains(strings.ToLower(cmd), strings.ToLower(p)) {
			return Check{IsDangerous: true, Pattern: p}
		}
	}
	for _, p := range warnings {
		if strings.Contains(strings.ToLower(cmd), strings.ToLower(p)) {
			return Check{IsWarning: true, Pattern: p}
		}
	}
	return Check{}
}
