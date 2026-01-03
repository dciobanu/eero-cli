package cmd

import (
	"fmt"
)

// Guest handles the guest network command
func (a *App) Guest(args []string) error {
	if len(args) == 0 {
		return a.GuestStatus()
	}

	switch args[0] {
	case "enable":
		return a.GuestEnable(true)
	case "disable":
		return a.GuestEnable(false)
	case "password":
		if len(args) < 2 {
			return fmt.Errorf("usage: guest password <new-password>")
		}
		return a.GuestPassword(args[1])
	default:
		return fmt.Errorf("unknown guest subcommand: %s", args[0])
	}
}

// GuestStatus shows the guest network status
func (a *App) GuestStatus() error {
	networkID, err := a.EnsureNetwork()
	if err != nil {
		return err
	}

	gn, err := a.Client.GetGuestNetwork(networkID)
	if err != nil {
		return fmt.Errorf("getting guest network: %w", err)
	}

	status := "disabled"
	if gn.Enabled {
		status = "enabled"
	}

	fmt.Println("Guest Network Status")
	fmt.Println("--------------------")
	fmt.Printf("Status:   %s\n", status)
	if gn.Name != "" {
		fmt.Printf("Name:     %s\n", gn.Name)
	}
	if gn.Enabled && gn.Password != "" {
		fmt.Printf("Password: %s\n", gn.Password)
	}

	return nil
}

// GuestEnable enables or disables the guest network
func (a *App) GuestEnable(enable bool) error {
	networkID, err := a.EnsureNetwork()
	if err != nil {
		return err
	}

	if err := a.Client.EnableGuestNetwork(networkID, enable); err != nil {
		return fmt.Errorf("updating guest network: %w", err)
	}

	action := "enabled"
	if !enable {
		action = "disabled"
	}
	fmt.Printf("Guest network has been %s\n", action)

	return nil
}

// GuestPassword sets the guest network password
func (a *App) GuestPassword(password string) error {
	networkID, err := a.EnsureNetwork()
	if err != nil {
		return err
	}

	if err := a.Client.SetGuestNetworkPassword(networkID, password); err != nil {
		return fmt.Errorf("updating guest network password: %w", err)
	}

	fmt.Println("Guest network password has been updated")

	return nil
}
