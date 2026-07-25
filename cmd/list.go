package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/corebrim/agentreg/internal/client"
	"github.com/spf13/cobra"
)

var listFormat string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered agents",
	RunE: func(_ *cobra.Command, _ []string) error {
		agents, err := client.New(registryURL).List(context.Background())
		if err != nil {
			return err
		}
		if listFormat == "json" {
			return printJSON(agents)
		}
		if len(agents) == 0 {
			fmt.Println("no agents registered")
			return nil
		}
		tw := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
		fmt.Fprintln(tw, "NAME\tCAPABILITIES\tSTATUS\tSOURCE\tLAST HEARTBEAT\tENDPOINT")
		for _, a := range agents {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
				a.Name,
				strings.Join(a.Capabilities, ","),
				a.Status,
				a.Source,
				humanizeTime(a.LastHeartbeat),
				a.Endpoint,
			)
		}
		return tw.Flush()
	},
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func humanizeTime(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	d := time.Since(t)
	switch {
	case d < time.Second:
		return "just now"
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	default:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
}

func init() {
	listCmd.Flags().StringVar(&listFormat, "format", "table", "output format: table or json")
	rootCmd.AddCommand(listCmd)
}
