package cmd

import (
	"os"
	"strings"
)

// color is enabled only when writing to a real terminal and NO_COLOR is unset,
// so piped/redirected output (scripts, CI) stays clean.
var color = colorEnabled()

func colorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	fi, err := os.Stdout.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiDim    = "\033[2m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiCyan   = "\033[36m"
)

// paint wraps s in an ANSI code (a no-op when color is disabled).
func paint(s, code string) string {
	if !color || s == "" {
		return s
	}
	return code + s + ansiReset
}

// statusCode maps a health status to its color.
func statusCode(status string) string {
	switch status {
	case "healthy":
		return ansiGreen
	case "unhealthy":
		return ansiRed
	default:
		return ansiYellow
	}
}

// styleStatus returns a colorized status word.
func styleStatus(status string) string {
	return paint(status, statusCode(status))
}

// pad right-pads s with spaces to width w (based on plain length).
func pad(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
