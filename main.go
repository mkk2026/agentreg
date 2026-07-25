// Command agentctl is the agentreg CLI and daemon.
package main

import (
	"fmt"
	"os"

	"github.com/mkk2026/agentreg/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
