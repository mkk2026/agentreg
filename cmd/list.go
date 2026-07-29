package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/mkk2026/agentreg/internal/client"
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
			fmt.Println(paint("no agents registered", ansiDim))
			return nil
		}

		headers := []string{"NAME", "CAPABILITIES", "STATUS", "SOURCE", "LABELS", "LAST HEARTBEAT", "ENDPOINT"}
		const statusCol = 2

		rows := make([][]string, len(agents))
		var healthy, unhealthy int
		for i, a := range agents {
			st := string(a.Status)
			switch st {
			case "healthy":
				healthy++
			case "unhealthy":
				unhealthy++
			}
			rows[i] = []string{
				a.Name,
				strings.Join(a.Capabilities, ","),
				st,
				a.Source,
				formatLabels(a.Labels),
				humanizeTime(a.LastHeartbeat),
				a.Endpoint,
			}
		}

		widths := make([]int, len(headers))
		for i, h := range headers {
			widths[i] = len(h)
		}
		for _, r := range rows {
			for i, c := range r {
				if len(c) > widths[i] {
					widths[i] = len(c)
				}
			}
		}

		fmt.Println(paint(joinCells(headers, widths, -1, ""), ansiBold))
		for _, r := range rows {
			fmt.Println(joinCells(r, widths, statusCol, statusCode(r[statusCol])))
		}

		summary := []string{paint(fmt.Sprintf("%d agent%s", len(agents), plural(len(agents))), ansiDim)}
		if healthy > 0 {
			summary = append(summary, paint(fmt.Sprintf("%d healthy", healthy), ansiGreen))
		}
		if unhealthy > 0 {
			summary = append(summary, paint(fmt.Sprintf("%d unhealthy", unhealthy), ansiRed))
		}
		fmt.Println(strings.Join(summary, paint(" · ", ansiDim)))
		return nil
	},
}

// joinCells pads each cell to its column width and joins with two spaces.
// If colorCol >= 0, that cell is wrapped in colorCode (padding happens first,
// so alignment is preserved regardless of the escape codes).
func joinCells(cells []string, widths []int, colorCol int, colorCode string) string {
	var b strings.Builder
	for i, c := range cells {
		cell := pad(c, widths[i])
		if i == colorCol {
			cell = paint(cell, colorCode)
		}
		b.WriteString(cell)
		if i < len(cells)-1 {
			b.WriteString("  ")
		}
	}
	return b.String()
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// formatLabels renders a labels map as sorted key=value pairs, or "-" if empty.
func formatLabels(m map[string]string) string {
	if len(m) == 0 {
		return "-"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+m[k])
	}
	return strings.Join(parts, ",")
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
