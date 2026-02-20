package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:   "gl",
	Short: "gate-limiter — API rate limiting reverse proxy",
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "config file path (default: $GATE_LIMITER_CONFIG or config.yml)")
}

func resolveConfigPath() string {
	if configPath != "" {
		return configPath
	}
	if env := os.Getenv("GATE_LIMITER_CONFIG"); env != "" {
		return env
	}
	return "config.yml"
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
