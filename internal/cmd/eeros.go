package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dorin/eero-cli/internal/api"
)

// Eeros handles the eeros command
func (a *App) Eeros(args []string) error {
	if len(args) == 0 {
		return a.ListEeros()
	}

	switch args[0] {
	case "list":
		return a.ListEeros()
	case "inspect":
		if len(args) < 2 {
			return fmt.Errorf("usage: eeros inspect <eero>")
		}
		return a.InspectEero(args[1])
	case "reboot":
		if len(args) < 2 {
			return fmt.Errorf("usage: eeros reboot <eero>")
		}
		return a.RebootEero(args[1])
	default:
		return fmt.Errorf("unknown eeros subcommand: %s", args[0])
	}
}

// ListEeros lists all eero nodes on the network
func (a *App) ListEeros() error {
	networkID, err := a.EnsureNetwork()
	if err != nil {
		return err
	}

	eeros, err := a.Client.GetEeros(networkID)
	if err != nil {
		return fmt.Errorf("getting eeros: %w", err)
	}

	if len(eeros) == 0 {
		fmt.Println("No eero nodes found")
		return nil
	}

	headers := []string{"ID", "LOCATION", "STATUS", "GATEWAY", "IP", "MODEL", "CLIENTS", "SIGNAL", "TYPE"}
	var rows [][]string

	for _, e := range eeros {
		eeroID := api.ExtractEeroID(e.URL)

		// Format status (lowercase to match devices output)
		status := strings.ToLower(e.State)

		// Gateway indicator
		gateway := "no"
		if e.Gateway {
			gateway = "yes"
		}

		// Format mesh quality as fraction
		signal := fmt.Sprintf("%d/5", e.MeshQualityBars)

		// Connection type
		connType := "wireless"
		if e.Wired {
			connType = "wired"
		}

		rows = append(rows, []string{
			eeroID,
			e.Location,
			status,
			gateway,
			e.IPAddress,
			e.Model,
			fmt.Sprintf("%d", e.ConnectedClientsCount),
			signal,
			connType,
		})
	}

	PrintTable(headers, rows)
	fmt.Printf("\nTotal: %d eero nodes\n", len(eeros))

	return nil
}

// findEeroID finds an eero by partial ID, serial, or location
func (a *App) findEeroID(networkID, query string) (string, error) {
	eeros, err := a.Client.GetEeros(networkID)
	if err != nil {
		return "", fmt.Errorf("getting eeros: %w", err)
	}

	query = strings.ToLower(query)

	for _, e := range eeros {
		eeroID := api.ExtractEeroID(e.URL)

		// Exact ID match
		if eeroID == query {
			return eeroID, nil
		}

		// Partial ID match
		if strings.HasPrefix(strings.ToLower(eeroID), query) {
			return eeroID, nil
		}

		// Serial match
		if strings.EqualFold(e.Serial, query) {
			return eeroID, nil
		}

		// Location match (case-insensitive contains)
		if strings.Contains(strings.ToLower(e.Location), query) {
			return eeroID, nil
		}
	}

	return "", fmt.Errorf("eero not found: %s", query)
}

// InspectEero prints the full eero state as JSON
func (a *App) InspectEero(eeroQuery string) error {
	networkID, err := a.EnsureNetwork()
	if err != nil {
		return err
	}

	eeroID, err := a.findEeroID(networkID, eeroQuery)
	if err != nil {
		return err
	}

	rawJSON, err := a.Client.GetEeroRaw(eeroID)
	if err != nil {
		return fmt.Errorf("getting eero: %w", err)
	}

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, rawJSON, "", "  "); err != nil {
		return fmt.Errorf("formatting JSON: %w", err)
	}

	fmt.Println(prettyJSON.String())

	return nil
}

// RebootEero reboots a single eero node
func (a *App) RebootEero(eeroQuery string) error {
	networkID, err := a.EnsureNetwork()
	if err != nil {
		return err
	}

	// Get eeros to find matching one and get its location for confirmation
	eeros, err := a.Client.GetEeros(networkID)
	if err != nil {
		return fmt.Errorf("getting eeros: %w", err)
	}

	eeroID, err := a.findEeroID(networkID, eeroQuery)
	if err != nil {
		return err
	}

	// Find the eero to get its location
	var location string
	for _, e := range eeros {
		if api.ExtractEeroID(e.URL) == eeroID {
			location = e.Location
			break
		}
	}

	if err := a.Client.RebootEero(eeroID); err != nil {
		return fmt.Errorf("rebooting eero: %w", err)
	}

	fmt.Printf("Rebooting eero %s (%s)...\n", eeroID, location)
	return nil
}
