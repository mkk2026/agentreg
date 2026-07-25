package cmd

import (
	"context"
	"fmt"

	"github.com/mkk2026/agentreg/internal/client"
	"github.com/spf13/cobra"
)

var findFormat string

var findCmd = &cobra.Command{
	Use:   "find <capability>",
	Short: "Find agents that provide a capability",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		agents, err := client.New(registryURL).Find(context.Background(), args[0])
		if err != nil {
			return err
		}
		if findFormat == "json" {
			return printJSON(agents)
		}
		if len(agents) == 0 {
			fmt.Printf("no agents provide capability %q\n", args[0])
			return nil
		}
		for _, a := range agents {
			fmt.Printf("%s\t%s\t%s\n", a.Name, a.Status, a.Endpoint)
		}
		return nil
	},
}

func init() {
	findCmd.Flags().StringVar(&findFormat, "format", "table", "output format: table or json")
	rootCmd.AddCommand(findCmd)
}
