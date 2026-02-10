package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dorin/eero-cli/internal/api"
)

// Reservations handles the reservations command
func (a *App) Reservations(args []string) error {
	if len(args) == 0 {
		return a.ListReservations()
	}

	switch args[0] {
	case "add":
		if len(args) < 3 {
			return fmt.Errorf("usage: reservations add <mac> <ip> [description]")
		}
		desc := ""
		if len(args) >= 4 {
			desc = strings.Join(args[3:], " ")
		}
		return a.AddReservation(args[1], args[2], desc)
	case "remove":
		if len(args) < 2 {
			return fmt.Errorf("usage: reservations remove <id|mac|ip>")
		}
		return a.RemoveReservation(args[1])
	case "inspect":
		if len(args) < 2 {
			return fmt.Errorf("usage: reservations inspect <id|mac|ip>")
		}
		return a.InspectReservation(args[1])
	default:
		return fmt.Errorf("unknown reservations subcommand: %s", args[0])
	}
}

// ListReservations lists all DHCP reservations
func (a *App) ListReservations() error {
	networkID, err := a.EnsureNetwork()
	if err != nil {
		return err
	}

	reservations, err := a.Client.GetReservations(networkID)
	if err != nil {
		return fmt.Errorf("getting reservations: %w", err)
	}

	headers := []string{"IP", "MAC", "DESCRIPTION", "ID"}
	var rows [][]string
	for _, r := range reservations {
		rows = append(rows, []string{
			r.IP,
			r.MAC,
			r.Description,
			api.ExtractReservationID(r.URL),
		})
	}

	PrintTable(headers, rows)
	return nil
}

// AddReservation creates a new DHCP reservation
func (a *App) AddReservation(mac, ip, description string) error {
	networkID, err := a.EnsureNetwork()
	if err != nil {
		return err
	}

	if err := a.Client.CreateReservation(networkID, ip, mac, description); err != nil {
		return fmt.Errorf("creating reservation: %w", err)
	}

	fmt.Printf("Reservation created: %s -> %s\n", mac, ip)
	return nil
}

// RemoveReservation deletes a DHCP reservation
func (a *App) RemoveReservation(query string) error {
	networkID, err := a.EnsureNetwork()
	if err != nil {
		return err
	}

	reservationID, err := a.findReservationID(networkID, query)
	if err != nil {
		return err
	}

	if err := a.Client.DeleteReservation(networkID, reservationID); err != nil {
		return fmt.Errorf("deleting reservation: %w", err)
	}

	fmt.Println("Reservation deleted")
	return nil
}

// InspectReservation shows the raw JSON for a reservation
func (a *App) InspectReservation(query string) error {
	networkID, err := a.EnsureNetwork()
	if err != nil {
		return err
	}

	reservationID, err := a.findReservationID(networkID, query)
	if err != nil {
		return err
	}

	data, err := a.Client.GetReservationRaw(networkID, reservationID)
	if err != nil {
		return fmt.Errorf("getting reservation: %w", err)
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, data, "", "  "); err != nil {
		return fmt.Errorf("formatting JSON: %w", err)
	}

	fmt.Println(pretty.String())
	return nil
}

// findReservationID resolves a query (ID, MAC, or IP) to a reservation ID
func (a *App) findReservationID(networkID, query string) (string, error) {
	reservations, err := a.Client.GetReservations(networkID)
	if err != nil {
		return "", fmt.Errorf("getting reservations: %w", err)
	}

	query = strings.ToLower(query)

	for _, r := range reservations {
		reservationID := api.ExtractReservationID(r.URL)

		// Exact ID match
		if reservationID == query {
			return reservationID, nil
		}

		// MAC match (normalized)
		if strings.ToLower(r.MAC) == query || strings.ReplaceAll(strings.ToLower(r.MAC), ":", "") == strings.ReplaceAll(query, ":", "") {
			return reservationID, nil
		}

		// IP match
		if r.IP == query {
			return reservationID, nil
		}
	}

	return "", fmt.Errorf("reservation not found: %s", query)
}
