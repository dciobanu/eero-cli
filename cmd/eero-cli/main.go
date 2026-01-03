package main

import (
	"fmt"
	"os"

	"github.com/dorin/eero-cli/internal/cmd"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	args := os.Args[1:]

	if len(args) == 0 {
		cmd.Usage()
		return nil
	}

	app, err := cmd.NewApp()
	if err != nil {
		return err
	}

	command := args[0]
	subArgs := args[1:]

	switch command {
	case "help", "-h", "--help":
		cmd.Usage()
		return nil

	case "login":
		return app.Login()

	case "logout":
		return app.Logout()

	case "status":
		return app.Status()

	case "devices":
		return app.Devices(subArgs)

	case "profiles":
		return app.Profiles(subArgs)

	case "guest":
		return app.Guest(subArgs)

	case "reboot":
		return app.Reboot()

	default:
		return fmt.Errorf("unknown command: %s\nRun 'eero-cli help' for usage", command)
	}
}
