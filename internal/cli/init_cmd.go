package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate a default config.yml in the current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		const filename = "config.yml"

		if _, err := os.Stat(filename); err == nil {
			return fmt.Errorf("%s already exists", filename)
		}

		if err := os.WriteFile(filename, []byte(defaultConfigTemplate), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}

		fmt.Printf("Created %s\n", filename)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
