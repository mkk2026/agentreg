package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/mkk2026/agentreg/internal/agent"
	"github.com/mkk2026/agentreg/internal/client"
	"github.com/spf13/cobra"
)

var (
	regCapabilities []string
	regEndpoint     string
)

var registerCmd = &cobra.Command{
	Use:   "register <name>",
	Short: "Register an agent with the registry",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		if regEndpoint == "" {
			return fmt.Errorf("--endpoint is required")
		}
		a := agent.Agent{
			Name:         args[0],
			Capabilities: normalizeCaps(regCapabilities),
			Endpoint:     regEndpoint,
			Source:       agent.SourceLocal,
		}
		stored, err := client.New(registryURL).Register(context.Background(), a)
		if err != nil {
			return err
		}
		fmt.Printf("%s registered %s %s → %s\n",
			paint("✓", ansiGreen),
			paint(stored.Name, ansiBold),
			paint("["+strings.Join(stored.Capabilities, ",")+"]", ansiDim),
			stored.Endpoint,
		)
		return nil
	},
}

// normalizeCaps splits comma-joined values and trims blanks so both
// "-c a,b" and "-c a -c b" behave the same.
func normalizeCaps(in []string) []string {
	var out []string
	for _, v := range in {
		for _, part := range strings.Split(v, ",") {
			if p := strings.TrimSpace(part); p != "" {
				out = append(out, p)
			}
		}
	}
	return out
}

func init() {
	registerCmd.Flags().StringSliceVarP(&regCapabilities, "capabilities", "c", nil,
		"comma-separated capabilities (e.g. search,db-read)")
	registerCmd.Flags().StringVarP(&regEndpoint, "endpoint", "e", "", "agent endpoint URL")
	rootCmd.AddCommand(registerCmd)
}
