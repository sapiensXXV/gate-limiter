package cli

import (
	"fmt"
	"gate-limiter/config/settings"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the configuration file",
	RunE: func(_ *cobra.Command, _ []string) error {
		cfgPath := resolveConfigPath()
		_, err := settings.ParseAndValidateConfig(cfgPath)
		if err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
		fmt.Println("configuration is valid")
		return nil
	},
}
