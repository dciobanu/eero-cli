package cmd

import (
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
