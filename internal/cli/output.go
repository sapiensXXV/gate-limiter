package cli

import (
	"os"

	"github.com/fatih/color"
)

var errorPrinter = color.New(color.FgRed, color.Bold)

func printError(err error) {
	errorPrinter.Fprintf(os.Stderr, "Error: %v\n", err)
}
