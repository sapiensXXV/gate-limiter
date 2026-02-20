package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "v0.2.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("gate-limiter %s\n", Version)
	},
}
