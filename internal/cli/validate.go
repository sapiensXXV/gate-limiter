package cli

import (
	"fmt"
	"gate-limiter/config/settings"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath := resolveConfigPath()

		config, err := settings.ParseAndValidateConfig(cfgPath)
		if err != nil {
			return fmt.Errorf("configuration invalid: %w", err)
		}

		fmt.Println("Configuration is valid.")
		fmt.Printf("  strategy : %s\n", config.RateLimiter.Strategy)
		fmt.Printf("  identity : %s\n", config.RateLimiter.Identity.Key)
		fmt.Printf("  APIs     : %d\n", len(config.RateLimiter.Apis))
		fmt.Printf("  port     : %d\n", config.RateLimiter.Port)
		fmt.Printf("  adminPort: %d\n", config.RateLimiter.AdminPort)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
