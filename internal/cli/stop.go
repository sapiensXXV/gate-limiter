package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

const pidFile = "gl.pid"

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the daemon process",
	RunE: func(_ *cobra.Command, _ []string) error {
		pid, err := readPidFile()
		if err != nil {
			return err
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			return fmt.Errorf("failed to find process %d: %w", pid, err)
		}

		if err := process.Signal(syscall.SIGTERM); err != nil {
			os.Remove(pidFile)
			return fmt.Errorf("failed to send SIGTERM to process %d: %w", pid, err)
		}

		os.Remove(pidFile)
		fmt.Printf("gate-limiter stopped (PID: %d)\n", pid)
		return nil
	},
}

func readPidFile() (int, error) {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("no daemon running (%s not found)", pidFile)
		}
		return 0, fmt.Errorf("failed to read %s: %w", pidFile, err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("invalid PID in %s: %w", pidFile, err)
	}
	return pid, nil
}
