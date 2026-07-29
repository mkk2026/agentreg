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
			fmt.Println(paint(fmt.Sprintf("no agents provide capability %q", args[0]), ansiDim))
			return nil
		}
		nameW := 0
		for _, a := range agents {
			if len(a.Name) > nameW {
				nameW = len(a.Name)
			}
		}
		for _, a := range agents {
			fmt.Printf("%s  %s  %s\n",
				paint(pad(a.Name, nameW), ansiBold),
				paint(pad(string(a.Status), 9), statusCode(string(a.Status))),
				paint(a.Endpoint, ansiDim),
			)
		}
		return nil
	},
}

func init() {
	findCmd.Flags().StringVar(&findFormat, "format", "table", "output format: table or json")
	rootCmd.AddCommand(findCmd)
}
