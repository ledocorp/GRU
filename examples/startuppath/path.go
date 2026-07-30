// Package startuppath resolves optional file paths from argv and environment.
package startuppath

import "strings"

// Resolve returns the first non-flag CLI argument, or envPath when argv is empty.
func Resolve(args []string, envPath string) string {
	if p := FirstFileArg(args); p != "" {
		return p
	}
	return strings.TrimSpace(envPath)
}

// FirstFileArg returns the first non-flag argument after the program name.
func FirstFileArg(args []string) string {
	if len(args) <= 1 {
		return ""
	}
	for _, arg := range args[1:] {
		if arg == "" || strings.HasPrefix(arg, "-") {
			continue
		}
		return strings.TrimSpace(strings.Trim(arg, `"`))
	}
	return ""
}
