// Package app wires the dictation services together and exposes the CLI
// command surface, exit codes and usage text.
package app

import (
	"fmt"
	"os"
)

const (
	exitOK    = 0
	exitErr   = 1
	exitUsage = 2
)

// Main dispatches args and returns the process exit code.
func Main(args []string) int {
	if len(args) == 0 {
		fmt.Print(usageText)
		return exitUsage
	}
	if cmd, ok := commands[args[0]]; ok {
		return execCommand(cmd, args[1:])
	}
	return runBuiltin(args)
}

func execCommand(cmd func([]string) error, args []string) int {
	if err := cmd(args); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return exitErr
	}
	return exitOK
}

func runBuiltin(args []string) int {
	switch args[0] {
	case "watchdog":
		RunWatchdog(args[1:])
		return exitOK
	case "-h", "--help", "help":
		fmt.Print(usageText)
		return exitOK
	default:
		fmt.Fprintln(os.Stderr, "unknown command:", args[0])
		fmt.Print(usageText)
		return exitUsage
	}
}
