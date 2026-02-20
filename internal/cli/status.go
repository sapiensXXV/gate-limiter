package cli

import (
	"fmt"
	"os"
	"syscall"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check whether the daemon is running",
	RunE: func(_ *cobra.Command, _ []string) error {
		pid, err := readPidFile()
		if err != nil {
			fmt.Println("gate-limiter is not running")
			return nil
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			fmt.Printf("gate-limiter is not running (stale %s, PID: %d)\n", pidFile, pid)
			return nil
		}

		// Signal 0 checks if the process exists without actually sending a signal
		if err := process.Signal(syscall.Signal(0)); err != nil {
			fmt.Printf("gate-limiter is not running (stale %s, PID: %d)\n", pidFile, pid)
			return nil
		}

		fmt.Printf("gate-limiter is running (PID: %d)\n", pid)
		return nil
	},
}
