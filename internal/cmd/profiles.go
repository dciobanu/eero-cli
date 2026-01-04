package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dorin/eero-cli/internal/api"
)

// Profiles handles the profiles command
func (a *App) Profiles(args []string) error {
	if len(args) == 0 {
		return a.ListProfiles()
	}

	switch args[0] {
	case "inspect":
		if len(args) < 2 {
			return fmt.Errorf("usage: profiles inspect <profile>")
		}
		return a.InspectProfile(args[1])
	case "pause":
		if len(args) < 2 {
			return fmt.Errorf("usage: profiles pause <profile-id>")
		}
		return a.PauseProfile(args[1], true)
	case "unpause":
		if len(args) < 2 {
			return fmt.Errorf("usage: profiles unpause <profile-id>")
		}
		return a.PauseProfile(args[1], false)
	case "add":
		if len(args) < 3 {
			return fmt.Errorf("usage: profiles add <profile> <device>")
		}
		return a.AddDeviceToProfile(args[1], args[2])
	case "remove":
		if len(args) < 3 {
			return fmt.Errorf("usage: profiles remove <profile> <device>")
		}
		return a.RemoveDeviceFromProfile(args[1], args[2])
	default:
		return fmt.Errorf("unknown profiles subcommand: %s", args[0])
	}
}

// ListProfiles lists all profiles on the network
func (a *App) ListProfiles() error {
	networkID, err := a.EnsureNetwork()
	if err != nil {
		return err
	}

	profiles, err := a.Client.GetProfiles(networkID)
	if err != nil {
		return fmt.Errorf("getting profiles: %w", err)
	}

	if len(profiles) == 0 {
		fmt.Println("No profiles configured")
		return nil
	}

	headers := []string{"ID", "NAME", "STATUS"}
	var rows [][]string

	for _, p := range profiles {
		status := "active"
		if p.Paused {
			status = "paused"
		}

		profileID := api.ExtractProfileID(p.URL)

		rows = append(rows, []string{
			profileID,
			p.Name,
			status,
		})
	}

	PrintTable(headers, rows)
	fmt.Printf("\nTotal: %d profiles\n", len(profiles))

	return nil
}

// findProfileID finds a profile by partial ID or name
func (a *App) findProfileID(networkID, query string) (string, error) {
	profiles, err := a.Client.GetProfiles(networkID)
	if err != nil {
		return "", fmt.Errorf("getting profiles: %w", err)
	}

	query = strings.ToLower(query)

	for _, p := range profiles {
		profileID := api.ExtractProfileID(p.URL)

		// Exact ID match
		if profileID == query {
			return profileID, nil
		}

		// Partial ID match
		if strings.HasPrefix(strings.ToLower(profileID), query) {
			return profileID, nil
		}

		// Name match
		if strings.EqualFold(p.Name, query) {
			return profileID, nil
		}
	}

	return "", fmt.Errorf("profile not found: %s", query)
}

// PauseProfile pauses or unpauses a profile
func (a *App) PauseProfile(profileQuery string, pause bool) error {
	networkID, err := a.EnsureNetwork()
	if err != nil {
		return err
	}

	profileID, err := a.findProfileID(networkID, profileQuery)
	if err != nil {
		return err
	}

	if err := a.Client.PauseProfile(networkID, profileID, pause); err != nil {
		return fmt.Errorf("updating profile: %w", err)
	}

	action := "paused"
	if !pause {
		action = "unpaused"
	}
	fmt.Printf("Profile %s has been %s\n", profileID, action)

	return nil
}

// InspectProfile prints the full profile state as JSON
func (a *App) InspectProfile(profileQuery string) error {
	networkID, err := a.EnsureNetwork()
	if err != nil {
		return err
	}

	profileID, err := a.findProfileID(networkID, profileQuery)
	if err != nil {
		return err
	}

	rawJSON, err := a.Client.GetProfileRaw(networkID, profileID)
	if err != nil {
		return fmt.Errorf("getting profile: %w", err)
	}

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, rawJSON, "", "  "); err != nil {
		return fmt.Errorf("formatting JSON: %w", err)
	}

	fmt.Println(prettyJSON.String())

	return nil
}

// AddDeviceToProfile adds a device to a profile
func (a *App) AddDeviceToProfile(profileQuery, deviceQuery string) error {
	networkID, err := a.EnsureNetwork()
	if err != nil {
		return err
	}

	profileID, err := a.findProfileID(networkID, profileQuery)
	if err != nil {
		return err
	}

	deviceID, err := a.findDeviceID(networkID, deviceQuery)
	if err != nil {
		return err
	}

	// Get current profile devices
	profile, err := a.Client.GetProfileDetails(networkID, profileID)
	if err != nil {
		return fmt.Errorf("getting profile: %w", err)
	}

	// Check if device is already in profile
	deviceURL := fmt.Sprintf("/2.2/networks/%s/devices/%s", networkID, deviceID)
	for _, d := range profile.Devices {
		if d.URL == deviceURL {
			return fmt.Errorf("device %s is already in profile %s", deviceID, profile.Name)
		}
	}

	// Add device to list
	deviceURLs := make([]string, len(profile.Devices)+1)
	for i, d := range profile.Devices {
		deviceURLs[i] = d.URL
	}
	deviceURLs[len(profile.Devices)] = deviceURL

	if err := a.Client.SetProfileDevices(networkID, profileID, deviceURLs); err != nil {
		return fmt.Errorf("updating profile: %w", err)
	}

	fmt.Printf("Device %s has been added to profile %s\n", deviceID, profile.Name)
	return nil
}

// RemoveDeviceFromProfile removes a device from a profile
func (a *App) RemoveDeviceFromProfile(profileQuery, deviceQuery string) error {
	networkID, err := a.EnsureNetwork()
	if err != nil {
		return err
	}

	profileID, err := a.findProfileID(networkID, profileQuery)
	if err != nil {
		return err
	}

	deviceID, err := a.findDeviceID(networkID, deviceQuery)
	if err != nil {
		return err
	}

	// Get current profile devices
	profile, err := a.Client.GetProfileDetails(networkID, profileID)
	if err != nil {
		return fmt.Errorf("getting profile: %w", err)
	}

	// Find and remove device from list
	deviceURL := fmt.Sprintf("/2.2/networks/%s/devices/%s", networkID, deviceID)
	found := false
	deviceURLs := make([]string, 0, len(profile.Devices))
	for _, d := range profile.Devices {
		if d.URL == deviceURL {
			found = true
		} else {
			deviceURLs = append(deviceURLs, d.URL)
		}
	}

	if !found {
		return fmt.Errorf("device %s is not in profile %s", deviceID, profile.Name)
	}

	if err := a.Client.SetProfileDevices(networkID, profileID, deviceURLs); err != nil {
		return fmt.Errorf("updating profile: %w", err)
	}

	fmt.Printf("Device %s has been removed from profile %s\n", deviceID, profile.Name)
	return nil
}
