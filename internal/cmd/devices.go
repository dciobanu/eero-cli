package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dorin/eero-cli/internal/api"
)

// DeviceFilters holds filter options for device listing
type DeviceFilters struct {
	Profile   string
	NoProfile bool
	Wired     bool
	Wireless  bool
	Online    bool
	Offline   bool
	Guest     bool
	NoGuest   bool
	Interval  int
}

// Devices handles the devices command
func (a *App) Devices(args []string) error {
	// Parse flags
	var filters DeviceFilters
	var filteredArgs []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--profile" && i+1 < len(args) {
			filters.Profile = args[i+1]
			i++ // skip the value
		} else if strings.HasPrefix(args[i], "--profile=") {
			filters.Profile = strings.TrimPrefix(args[i], "--profile=")
		} else if args[i] == "--wired" {
			filters.Wired = true
		} else if args[i] == "--wireless" {
			filters.Wireless = true
		} else if args[i] == "--online" {
			filters.Online = true
		} else if args[i] == "--offline" {
			filters.Offline = true
		} else if args[i] == "--guest" {
			filters.Guest = true
		} else if args[i] == "--noguest" {
			filters.NoGuest = true
		} else if args[i] == "--noprofile" {
			filters.NoProfile = true
		} else if args[i] == "--interval" && i+1 < len(args) {
			if v, err := strconv.Atoi(args[i+1]); err == nil {
				filters.Interval = v
			}
			i++ // skip the value
		} else if strings.HasPrefix(args[i], "--interval=") {
			if v, err := strconv.Atoi(strings.TrimPrefix(args[i], "--interval=")); err == nil {
				filters.Interval = v
			}
		} else {
			filteredArgs = append(filteredArgs, args[i])
		}
	}

	if len(filteredArgs) == 0 {
		return a.ListDevices(filters)
	}

	switch filteredArgs[0] {
	case "monitor":
		return a.MonitorDevices(filters)
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

// ListDevices lists all devices on the network, optionally filtered
func (a *App) ListDevices(filters DeviceFilters) error {
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
	if filters.Profile != "" {
		profiles, err := a.Client.GetProfiles(networkID)
		if err == nil {
			for _, p := range profiles {
				profileID := api.ExtractProfileID(p.URL)
				// Check if filter matches ID or name
				if strings.EqualFold(profileID, filters.Profile) || strings.EqualFold(p.Name, filters.Profile) {
					resolvedProfileName = p.Name
					resolvedProfileID = profileID
					break
				}
			}
		}
		if resolvedProfileName == "" {
			// No exact match found, use filter as-is for name matching
			resolvedProfileName = filters.Profile
		}
	}

	headers := []string{"ID", "NAME", "IP", "MAC", "STATUS", "TYPE", "PROFILE"}
	var rows [][]string
	var filteredCount int

	for _, d := range devices {
		profileDisplay := ""
		profileName := ""
		profileID := ""
		if d.IsGuest {
			profileDisplay = "Guest"
		} else if d.Profile != nil {
			profileName = d.Profile.Name
			profileID = api.ExtractProfileID(d.Profile.URL)
			profileDisplay = fmt.Sprintf("%s (%s)", profileName, profileID)
		}

		// Apply profile filter if specified (match by name or ID)
		if filters.Profile != "" {
			match := strings.EqualFold(profileName, resolvedProfileName) ||
				strings.EqualFold(profileID, filters.Profile)
			if !match {
				continue
			}
		}

		// Apply wired/wireless filter
		if filters.Wired && d.Wireless {
			continue
		}
		if filters.Wireless && !d.Wireless {
			continue
		}

		// Apply online/offline filter
		if filters.Online && !d.Connected {
			continue
		}
		if filters.Offline && d.Connected {
			continue
		}

		// Apply guest filter
		if filters.Guest && !d.IsGuest {
			continue
		}

		// Apply noguest filter
		if filters.NoGuest && d.IsGuest {
			continue
		}

		// Apply noprofile filter (no profile assigned, includes guests)
		if filters.NoProfile && d.Profile != nil {
			continue
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

	// Build filter description
	var filterParts []string
	if filters.Profile != "" {
		if resolvedProfileID != "" {
			filterParts = append(filterParts, fmt.Sprintf("profile: %s [%s]", resolvedProfileName, resolvedProfileID))
		} else {
			filterParts = append(filterParts, fmt.Sprintf("profile: %s", filters.Profile))
		}
	}
	if filters.Wired {
		filterParts = append(filterParts, "wired")
	}
	if filters.Wireless {
		filterParts = append(filterParts, "wireless")
	}
	if filters.Online {
		filterParts = append(filterParts, "online")
	}
	if filters.Offline {
		filterParts = append(filterParts, "offline")
	}
	if filters.Guest {
		filterParts = append(filterParts, "guest")
	}
	if filters.NoGuest {
		filterParts = append(filterParts, "no guest")
	}
	if filters.NoProfile {
		filterParts = append(filterParts, "no profile")
	}

	if len(filterParts) > 0 {
		fmt.Printf("\nTotal: %d devices (filtered by %s)\n", filteredCount, strings.Join(filterParts, ", "))
	} else {
		fmt.Printf("\nTotal: %d devices\n", len(devices))
	}

	return nil
}

// DeviceState tracks the state of a device for monitoring
type DeviceState struct {
	Name      string
	IP        string
	MAC       string
	Connected bool
	Paused    bool
	Blocked   bool
	Wireless  bool
	IsGuest   bool
	Profile   string
}

const (
	boldStart = "\033[1m"
	boldEnd   = "\033[0m"
)

// bold wraps text in bold escape codes
func bold(s string) string {
	return boldStart + s + boldEnd
}

// boldIf wraps text in bold if condition is true
func boldIf(s string, condition bool) string {
	if condition {
		return bold(s)
	}
	return s
}

// MonitorDevices monitors devices for state changes
func (a *App) MonitorDevices(filters DeviceFilters) error {
	networkID, err := a.EnsureNetwork()
	if err != nil {
		return err
	}

	interval := filters.Interval
	if interval <= 0 {
		interval = 10
	}

	// Resolve profile filter once
	var resolvedProfileName string
	if filters.Profile != "" {
		profiles, err := a.Client.GetProfiles(networkID)
		if err == nil {
			for _, p := range profiles {
				profileID := api.ExtractProfileID(p.URL)
				if strings.EqualFold(profileID, filters.Profile) || strings.EqualFold(p.Name, filters.Profile) {
					resolvedProfileName = p.Name
					break
				}
			}
		}
		if resolvedProfileName == "" {
			resolvedProfileName = filters.Profile
		}
	}

	fmt.Printf("Monitoring devices every %d seconds. Press Ctrl+C to stop.\n\n", interval)

	// Print table header
	printMonitorHeader()

	// Track previous state
	prevState := make(map[string]DeviceState)
	first := true

	for {
		devices, err := a.Client.GetDevices(networkID)
		if err != nil {
			fmt.Printf("[%s] Error fetching devices: %v\n", time.Now().Format("15:04:05"), err)
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		}

		for _, d := range devices {
			// Apply filters
			profileName := ""
			profileDisplay := ""
			if d.IsGuest {
				profileDisplay = "Guest"
			} else if d.Profile != nil {
				profileName = d.Profile.Name
				profileID := api.ExtractProfileID(d.Profile.URL)
				profileDisplay = fmt.Sprintf("%s (%s)", profileName, profileID)
			}

			if filters.Profile != "" {
				profileID := ""
				if d.Profile != nil {
					profileID = api.ExtractProfileID(d.Profile.URL)
				}
				match := strings.EqualFold(profileName, resolvedProfileName) ||
					strings.EqualFold(profileID, filters.Profile)
				if !match {
					continue
				}
			}

			if filters.Wired && d.Wireless {
				continue
			}
			if filters.Wireless && !d.Wireless {
				continue
			}
			if filters.Online && !d.Connected {
				continue
			}
			if filters.Offline && d.Connected {
				continue
			}
			if filters.Guest && !d.IsGuest {
				continue
			}
			if filters.NoGuest && d.IsGuest {
				continue
			}
			if filters.NoProfile && d.Profile != nil {
				continue
			}

			deviceID := api.ExtractDeviceID(d.URL)
			currentState := DeviceState{
				Name:      d.DisplayName(),
				IP:        d.IP,
				MAC:       d.MAC,
				Connected: d.Connected,
				Paused:    d.Paused,
				Blocked:   d.Blocked,
				Wireless:  d.Wireless,
				IsGuest:   d.IsGuest,
				Profile:   profileDisplay,
			}

			prev, exists := prevState[deviceID]
			hasChanges := false

			if !first && exists {
				// Check for any changes
				hasChanges = prev.Connected != currentState.Connected ||
					prev.Paused != currentState.Paused ||
					prev.Blocked != currentState.Blocked ||
					prev.IP != currentState.IP
			} else if !first && !exists {
				// New device
				hasChanges = true
			}

			if hasChanges {
				printMonitorRow(deviceID, prev, currentState, !exists)
			}

			prevState[deviceID] = currentState
		}

		first = false
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func printMonitorHeader() {
	fmt.Printf("%-8s  %-12s  %-25s  %-15s  %-17s  %-7s  %-8s  %s\n",
		"TIME", "ID", "NAME", "IP", "MAC", "STATUS", "TYPE", "PROFILE")
	fmt.Printf("%-8s  %-12s  %-25s  %-15s  %-17s  %-7s  %-8s  %s\n",
		"--------", "------------", "-------------------------", "---------------", "-----------------", "-------", "--------", "------------------------")
}

// pad pads a string to the given width
func pad(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func printMonitorRow(deviceID string, prev, curr DeviceState, isNew bool) {
	timestamp := time.Now().Format("15:04:05")

	// Determine status
	status := "offline"
	if curr.Connected {
		status = "online"
	}
	if curr.Paused {
		status = "paused"
	}
	if curr.Blocked {
		status = "blocked"
	}

	connType := "wired"
	if curr.Wireless {
		connType = "wireless"
	}

	// Pad values first, then apply bold to preserve alignment
	name := pad(curr.Name, 25)
	ip := pad(curr.IP, 15)
	mac := pad(curr.MAC, 17)
	statusPad := pad(status, 7)
	connTypePad := pad(connType, 8)

	if isNew {
		// New device - bold everything
		name = bold(name)
		ip = bold(ip)
		statusPad = bold(statusPad)
	} else {
		// Bold only changed values
		statusChanged := prev.Connected != curr.Connected || prev.Paused != curr.Paused || prev.Blocked != curr.Blocked
		statusPad = boldIf(statusPad, statusChanged)
		ip = boldIf(ip, prev.IP != curr.IP)
	}

	fmt.Printf("%-8s  %-12s  %s  %s  %s  %s  %s  %s\n",
		timestamp, deviceID, name, ip, mac, statusPad, connTypePad, curr.Profile)
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
