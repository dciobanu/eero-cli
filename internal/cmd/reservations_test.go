package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/dorin/eero-cli/internal/api"
)

func testReservations() []api.Reservation {
	return []api.Reservation{
		{URL: "/2.2/networks/12345/reservations/res1", IP: "192.168.1.10", MAC: "11:22:33:44:55:66", Description: "NAS Server"},
		{URL: "/2.2/networks/12345/reservations/res2", IP: "192.168.1.20", MAC: "AA:BB:CC:DD:EE:FF", Description: "Printer"},
	}
}

func TestListReservations(t *testing.T) {
	mock := &mockClient{
		GetReservationsFn: func(networkID string) ([]api.Reservation, error) {
			return testReservations(), nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.ListReservations(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "192.168.1.10") {
		t.Error("output missing IP 192.168.1.10")
	}
	if !strings.Contains(out, "NAS Server") {
		t.Error("output missing description 'NAS Server'")
	}
	if !strings.Contains(out, "res1") {
		t.Error("output missing reservation ID res1")
	}
}

func TestAddReservation(t *testing.T) {
	var gotIP, gotMAC, gotDesc string
	mock := &mockClient{
		CreateReservationFn: func(networkID, ip, mac, description string) error {
			gotIP = ip
			gotMAC = mac
			gotDesc = description
			return nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.AddReservation("AA:BB:CC:DD:EE:FF", "192.168.1.50", "Test Device"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if gotIP != "192.168.1.50" {
		t.Errorf("IP = %q, want %q", gotIP, "192.168.1.50")
	}
	if gotMAC != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("MAC = %q, want %q", gotMAC, "AA:BB:CC:DD:EE:FF")
	}
	if gotDesc != "Test Device" {
		t.Errorf("Description = %q, want %q", gotDesc, "Test Device")
	}
	if !strings.Contains(out, "Reservation created") {
		t.Error("output missing confirmation message")
	}
}

func TestRemoveReservation(t *testing.T) {
	var deletedID string
	mock := &mockClient{
		GetReservationsFn: func(networkID string) ([]api.Reservation, error) {
			return testReservations(), nil
		},
		DeleteReservationFn: func(networkID, reservationID string) error {
			deletedID = reservationID
			return nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.RemoveReservation("res1"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if deletedID != "res1" {
		t.Errorf("deleted = %q, want %q", deletedID, "res1")
	}
	if !strings.Contains(out, "Reservation deleted") {
		t.Error("output missing confirmation")
	}
}

func TestInspectReservation(t *testing.T) {
	mock := &mockClient{
		GetReservationsFn: func(networkID string) ([]api.Reservation, error) {
			return testReservations(), nil
		},
		GetReservationRawFn: func(networkID, reservationID string) (json.RawMessage, error) {
			return json.RawMessage(`{"ip":"192.168.1.10","mac":"11:22:33:44:55:66"}`), nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.InspectReservation("res1"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "192.168.1.10") {
		t.Error("output missing IP in JSON")
	}
}

func TestFindReservationByMAC(t *testing.T) {
	mock := &mockClient{
		GetReservationsFn: func(networkID string) ([]api.Reservation, error) {
			return testReservations(), nil
		},
		DeleteReservationFn: func(networkID, reservationID string) error {
			return nil
		},
	}
	app := newTestApp(mock)

	// Find by MAC (case-insensitive)
	captureStdout(t, func() {
		if err := app.RemoveReservation("aa:bb:cc:dd:ee:ff"); err != nil {
			t.Fatalf("find by MAC failed: %v", err)
		}
	})
}

func TestFindReservationByIP(t *testing.T) {
	mock := &mockClient{
		GetReservationsFn: func(networkID string) ([]api.Reservation, error) {
			return testReservations(), nil
		},
		DeleteReservationFn: func(networkID, reservationID string) error {
			return nil
		},
	}
	app := newTestApp(mock)

	captureStdout(t, func() {
		if err := app.RemoveReservation("192.168.1.20"); err != nil {
			t.Fatalf("find by IP failed: %v", err)
		}
	})
}

func TestFindReservationByMACWithoutColons(t *testing.T) {
	mock := &mockClient{
		GetReservationsFn: func(networkID string) ([]api.Reservation, error) {
			return testReservations(), nil
		},
		DeleteReservationFn: func(networkID, reservationID string) error {
			return nil
		},
	}
	app := newTestApp(mock)

	captureStdout(t, func() {
		if err := app.RemoveReservation("112233445566"); err != nil {
			t.Fatalf("find by MAC without colons failed: %v", err)
		}
	})
}

func TestFindReservationNotFound(t *testing.T) {
	mock := &mockClient{
		GetReservationsFn: func(networkID string) ([]api.Reservation, error) {
			return testReservations(), nil
		},
	}
	app := newTestApp(mock)

	err := app.RemoveReservation("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent reservation")
	}
	if !strings.Contains(err.Error(), "reservation not found") {
		t.Errorf("error = %q, want 'reservation not found'", err.Error())
	}
}
