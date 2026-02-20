package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initConfigCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a default config.yml",
	RunE: func(_ *cobra.Command, _ []string) error {
		if _, err := os.Stat("config.yml"); err == nil {
			return fmt.Errorf("config.yml already exists")
		}
		if err := os.WriteFile("config.yml", []byte(defaultConfigTemplate), 0644); err != nil {
			return fmt.Errorf("failed to create config.yml: %w", err)
		}
		fmt.Println("config.yml created")
		return nil
	},
}
