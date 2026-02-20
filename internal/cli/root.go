package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:           "gl",
	Short:         "gate-limiter — API rate limiting reverse proxy",
	Long:          "gate-limiter is a configurable rate-limiting reverse proxy.\nIt supports token bucket, leaky bucket, fixed window counter,\nsliding window log, and sliding window counter algorithms.",
	SilenceErrors: true,
	SilenceUsage:  true,
	Example: `  gl run                          # Start server (foreground)
  gl run -d                        # Start server (daemon mode)
  gl run -c /etc/gl/config.yml     # Start with specific config
  gl validate                      # Validate config file
  gl init                          # Generate default config.yml
  gl stop                          # Stop daemon process
  gl status                        # Check daemon status`,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "config file path (default: $GATE_LIMITER_CONFIG or ./config.yml)")

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(initConfigCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(statusCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		printError(err)
		os.Exit(1)
	}
}
