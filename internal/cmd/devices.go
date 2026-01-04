package cmd

import (
	"fmt"
	"strings"

	"github.com/dorin/eero-cli/internal/api"
)

// Devices handles the devices command
func (a *App) Devices(args []string) error {
	// Parse --profile flag
	var profileFilter string
	var filteredArgs []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--profile" && i+1 < len(args) {
			profileFilter = args[i+1]
			i++ // skip the value
		} else if strings.HasPrefix(args[i], "--profile=") {
			profileFilter = strings.TrimPrefix(args[i], "--profile=")
		} else {
			filteredArgs = append(filteredArgs, args[i])
		}
	}

	if len(filteredArgs) == 0 {
		return a.ListDevices(profileFilter)
	}

	switch filteredArgs[0] {
	case "pause":
		if len(filteredArgs) < 2 {
			return fmt.Errorf("usage: devices pause <device-id>")
		}
		return a.PauseDevice(filteredArgs[1], true)
	case "unpause":
		if len(filteredArgs) < 2 {
			return fmt.Errorf("usage: devices unpause <device-id>")
		}
		return a.PauseDevice(filteredArgs[1], false)
	case "block":
		if len(filteredArgs) < 2 {
			return fmt.Errorf("usage: devices block <device-id>")
		}
		return a.BlockDevice(filteredArgs[1], true)
	case "unblock":
		if len(filteredArgs) < 2 {
			return fmt.Errorf("usage: devices unblock <device-id>")
		}
		return a.BlockDevice(filteredArgs[1], false)
	case "rename":
		if len(filteredArgs) < 3 {
			return fmt.Errorf("usage: devices rename <device-id> <name>")
		}
		return a.RenameDevice(filteredArgs[1], strings.Join(filteredArgs[2:], " "))
	default:
		return fmt.Errorf("unknown devices subcommand: %s", filteredArgs[0])
	}
}

// ListDevices lists all devices on the network, optionally filtered by profile
func (a *App) ListDevices(profileFilter string) error {
	networkID, err := a.EnsureNetwork()
	if err != nil {
		return err
	}

	devices, err := a.Client.GetDevices(networkID)
	if err != nil {
		return fmt.Errorf("getting devices: %w", err)
	}

	// Build profile ID to name map for resolving filter
	var resolvedProfileName string
	var resolvedProfileID string
	if profileFilter != "" {
		profiles, err := a.Client.GetProfiles(networkID)
		if err == nil {
			for _, p := range profiles {
				profileID := api.ExtractProfileID(p.URL)
				// Check if filter matches ID or name
				if strings.EqualFold(profileID, profileFilter) || strings.EqualFold(p.Name, profileFilter) {
					resolvedProfileName = p.Name
					resolvedProfileID = profileID
					break
				}
			}
		}
		if resolvedProfileName == "" {
			// No exact match found, use filter as-is for name matching
			resolvedProfileName = profileFilter
		}
	}

	headers := []string{"ID", "NAME", "IP", "MAC", "STATUS", "TYPE", "PROFILE"}
	var rows [][]string
	var filteredCount int

	for _, d := range devices {
		profileDisplay := ""
		profileName := ""
		profileID := ""
		if d.Profile != nil {
			profileName = d.Profile.Name
			profileID = api.ExtractProfileID(d.Profile.URL)
			profileDisplay = fmt.Sprintf("%s (%s)", profileName, profileID)
		}

		// Apply profile filter if specified (match by name or ID)
		if profileFilter != "" {
			match := strings.EqualFold(profileName, resolvedProfileName) ||
				strings.EqualFold(profileID, profileFilter)
			if !match {
				continue
			}
		}
		filteredCount++

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
			profileDisplay,
		})
	}

	PrintTable(headers, rows)
	if profileFilter != "" {
		if resolvedProfileID != "" {
			fmt.Printf("\nTotal: %d devices (filtered by profile: %s [%s])\n", filteredCount, resolvedProfileName, resolvedProfileID)
		} else {
			fmt.Printf("\nTotal: %d devices (filtered by profile: %s)\n", filteredCount, profileFilter)
		}
	} else {
		fmt.Printf("\nTotal: %d devices\n", len(devices))
	}

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
