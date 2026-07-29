// Package cmd holds the agentctl CLI commands.
package cmd

import (
	"runtime/debug"

	"github.com/spf13/cobra"
)

var registryURL string

// version is overridden at build time via -ldflags "-X .../cmd.version=...".
// When unset (e.g. `go install`), it falls back to the module version embedded
// in the binary's build info. See resolveVersion.
var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "agentctl",
	Short: "agentreg — DNS for AI agents: register, discover, and health-check agents",
	Long: `agentreg is a self-hosted registry for AI agents (MCP servers).

Consul tells you where a service is. agentreg tells you what an agent can do,
whether it's healthy, and (soon) how much to trust it.

Run the daemon with 'agentctl serve', then register and discover agents.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// resolveVersion prefers the ldflags-injected version (release builds); when
// that's absent it reads the module version baked in by `go install`.
func resolveVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return version
}

func init() {
	rootCmd.Version = resolveVersion()
	rootCmd.PersistentFlags().StringVar(&registryURL, "registry", "http://localhost:8080",
		"base URL of the agentreg daemon")
}
