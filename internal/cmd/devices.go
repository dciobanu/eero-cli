package cmd

import (
	"fmt"
	"strings"

	"github.com/dorin/eero-cli/internal/api"
)

// Devices handles the devices command
func (a *App) Devices(args []string) error {
	if len(args) == 0 {
		return a.ListDevices()
	}

	switch args[0] {
	case "pause":
		if len(args) < 2 {
			return fmt.Errorf("usage: devices pause <device-id>")
		}
		return a.PauseDevice(args[1], true)
	case "unpause":
		if len(args) < 2 {
			return fmt.Errorf("usage: devices unpause <device-id>")
		}
		return a.PauseDevice(args[1], false)
	case "block":
		if len(args) < 2 {
			return fmt.Errorf("usage: devices block <device-id>")
		}
		return a.BlockDevice(args[1], true)
	case "unblock":
		if len(args) < 2 {
			return fmt.Errorf("usage: devices unblock <device-id>")
		}
		return a.BlockDevice(args[1], false)
	case "rename":
		if len(args) < 3 {
			return fmt.Errorf("usage: devices rename <device-id> <name>")
		}
		return a.RenameDevice(args[1], strings.Join(args[2:], " "))
	default:
		return fmt.Errorf("unknown devices subcommand: %s", args[0])
	}
}

// ListDevices lists all devices on the network
func (a *App) ListDevices() error {
	networkID, err := a.EnsureNetwork()
	if err != nil {
		return err
	}

	devices, err := a.Client.GetDevices(networkID)
	if err != nil {
		return fmt.Errorf("getting devices: %w", err)
	}

	headers := []string{"ID", "NAME", "IP", "MAC", "STATUS", "TYPE"}
	var rows [][]string

	for _, d := range devices {
		status := "offline"
		if d.Connected {
			status = "online"
		}
		if d.Paused {
			status = "paused"
		}
		if d.Blocked {
			status = "blocked"
		}

		connType := "wired"
		if d.Wireless {
			connType = "wireless"
		}

		deviceID := api.ExtractDeviceID(d.URL)

		rows = append(rows, []string{
			deviceID,
			d.DisplayName(),
			d.IP,
			d.MAC,
			status,
			connType,
		})
	}

	PrintTable(headers, rows)
	fmt.Printf("\nTotal: %d devices\n", len(devices))

	return nil
}

// findDeviceID finds a device by partial ID, MAC, or name
func (a *App) findDeviceID(networkID, query string) (string, error) {
	devices, err := a.Client.GetDevices(networkID)
	if err != nil {
		return "", fmt.Errorf("getting devices: %w", err)
	}

	query = strings.ToLower(query)

	for _, d := range devices {
		deviceID := api.ExtractDeviceID(d.URL)

		// Exact ID match
		if deviceID == query {
			return deviceID, nil
		}

		// Partial ID match
		if strings.HasPrefix(strings.ToLower(deviceID), query) {
			return deviceID, nil
		}

		// MAC match
		if strings.ToLower(d.MAC) == query || strings.ReplaceAll(strings.ToLower(d.MAC), ":", "") == strings.ReplaceAll(query, ":", "") {
			return deviceID, nil
		}

		// Name match
		if strings.EqualFold(d.DisplayName(), query) {
			return deviceID, nil
		}
	}

	return "", fmt.Errorf("device not found: %s", query)
}

// PauseDevice pauses or unpauses a device
func (a *App) PauseDevice(deviceQuery string, pause bool) error {
	networkID, err := a.EnsureNetwork()
	if err != nil {
		return err
	}

	deviceID, err := a.findDeviceID(networkID, deviceQuery)
	if err != nil {
		return err
	}

	if err := a.Client.PauseDevice(networkID, deviceID, pause); err != nil {
		return fmt.Errorf("updating device: %w", err)
	}

	action := "paused"
	if !pause {
		action = "unpaused"
	}
	fmt.Printf("Device %s has been %s\n", deviceID, action)

	return nil
}

// BlockDevice blocks or unblocks a device
func (a *App) BlockDevice(deviceQuery string, block bool) error {
	networkID, err := a.EnsureNetwork()
	if err != nil {
		return err
	}

	deviceID, err := a.findDeviceID(networkID, deviceQuery)
	if err != nil {
		return err
	}

	if err := a.Client.BlockDevice(networkID, deviceID, block); err != nil {
		return fmt.Errorf("updating device: %w", err)
	}

	action := "blocked"
	if !block {
		action = "unblocked"
	}
	fmt.Printf("Device %s has been %s\n", deviceID, action)

	return nil
}

// RenameDevice sets a device's nickname
func (a *App) RenameDevice(deviceQuery, name string) error {
	networkID, err := a.EnsureNetwork()
	if err != nil {
		return err
	}

	deviceID, err := a.findDeviceID(networkID, deviceQuery)
	if err != nil {
		return err
	}

	if err := a.Client.SetDeviceNickname(networkID, deviceID, name); err != nil {
		return fmt.Errorf("updating device: %w", err)
	}

	fmt.Printf("Device %s has been renamed to '%s'\n", deviceID, name)

	return nil
}
