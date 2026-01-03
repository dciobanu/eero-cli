package cmd

import (
	"fmt"
)

// Reboot handles the reboot command
func (a *App) Reboot() error {
	networkID, err := a.EnsureNetwork()
	if err != nil {
		return err
	}

	if !Confirm("Are you sure you want to reboot the network? This will disconnect all devices temporarily.") {
		fmt.Println("Reboot cancelled")
		return nil
	}

	fmt.Println("Rebooting network...")

	if err := a.Client.Reboot(networkID); err != nil {
		return fmt.Errorf("rebooting network: %w", err)
	}

	fmt.Println("Network reboot initiated. Devices will reconnect automatically.")

	return nil
}
