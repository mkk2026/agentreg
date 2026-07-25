// Package cmd holds the agentctl CLI commands.
package cmd

import "github.com/spf13/cobra"

var registryURL string

// version is overridden at build time via -ldflags "-X .../cmd.version=...".
var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "agentctl",
	Short: "agentreg — DNS for AI agents: register, discover, and health-check agents",
	Long: `agentreg is a self-hosted registry for AI agents (MCP servers).

Consul tells you where a service is. agentreg tells you what an agent can do,
whether it's healthy, and (soon) how much to trust it.

Run the daemon with 'agentctl serve', then register and discover agents.`,
	Version:       version,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&registryURL, "registry", "http://localhost:8080",
		"base URL of the agentreg daemon")
}
