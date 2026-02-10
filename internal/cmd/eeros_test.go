package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/dorin/eero-cli/internal/api"
)

func testEeros() []api.Eero {
	return []api.Eero{
		{
			URL:                   "/2.2/eeros/8318690",
			Serial:                "SN12345678",
			Location:              "Living Room",
			Gateway:               true,
			IPAddress:             "192.168.1.1",
			Status:                "green",
			Model:                 "eero Pro 6E",
			OSVersion:             "7.2.1",
			Wired:                 true,
			State:                 "connected",
			MeshQualityBars:       5,
			ConnectedClientsCount: 12,
			HeartbeatOK:           true,
			IsPrimaryNode:         true,
			ConnectionType:        "wired",
		},
		{
			URL:                   "/2.2/eeros/8318691",
			Serial:                "SN87654321",
			Location:              "Bedroom",
			Gateway:               false,
			IPAddress:             "192.168.1.2",
			Status:                "green",
			Model:                 "eero 6+",
			OSVersion:             "7.2.1",
			Wired:                 false,
			State:                 "connected",
			MeshQualityBars:       3,
			ConnectedClientsCount: 5,
			HeartbeatOK:           true,
			IsPrimaryNode:         false,
			ConnectionType:        "wireless",
		},
	}
}

func TestListEeros(t *testing.T) {
	mock := &mockClient{
		GetEerosFn: func(networkID string) ([]api.Eero, error) {
			return testEeros(), nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.ListEeros(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "Living Room") {
		t.Error("output missing 'Living Room'")
	}
	if !strings.Contains(out, "Bedroom") {
		t.Error("output missing 'Bedroom'")
	}
	if !strings.Contains(out, "8318690") {
		t.Error("output missing eero ID")
	}
	if !strings.Contains(out, "eero Pro 6E") {
		t.Error("output missing model")
	}
	if !strings.Contains(out, "Total: 2 eero nodes") {
		t.Errorf("output missing total count, got:\n%s", out)
	}
}

func TestListEerosEmpty(t *testing.T) {
	mock := &mockClient{
		GetEerosFn: func(networkID string) ([]api.Eero, error) {
			return []api.Eero{}, nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.ListEeros(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "No eero nodes found") {
		t.Errorf("expected empty message, got:\n%s", out)
	}
}

func TestFindEeroByID(t *testing.T) {
	mock := &mockClient{
		GetEerosFn: func(networkID string) ([]api.Eero, error) {
			return testEeros(), nil
		},
	}
	app := newTestApp(mock)

	id, err := app.findEeroID("12345", "8318690")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "8318690" {
		t.Errorf("id = %q, want %q", id, "8318690")
	}
}

func TestFindEeroByPartialID(t *testing.T) {
	mock := &mockClient{
		GetEerosFn: func(networkID string) ([]api.Eero, error) {
			return testEeros(), nil
		},
	}
	app := newTestApp(mock)

	id, err := app.findEeroID("12345", "831869")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "8318690" {
		t.Errorf("id = %q, want %q", id, "8318690")
	}
}

func TestFindEeroBySerial(t *testing.T) {
	mock := &mockClient{
		GetEerosFn: func(networkID string) ([]api.Eero, error) {
			return testEeros(), nil
		},
	}
	app := newTestApp(mock)

	id, err := app.findEeroID("12345", "SN12345678")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "8318690" {
		t.Errorf("id = %q, want %q", id, "8318690")
	}
}

func TestFindEeroByLocation(t *testing.T) {
	mock := &mockClient{
		GetEerosFn: func(networkID string) ([]api.Eero, error) {
			return testEeros(), nil
		},
	}
	app := newTestApp(mock)

	id, err := app.findEeroID("12345", "bedroom")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "8318691" {
		t.Errorf("id = %q, want %q", id, "8318691")
	}
}

func TestFindEeroNotFound(t *testing.T) {
	mock := &mockClient{
		GetEerosFn: func(networkID string) ([]api.Eero, error) {
			return testEeros(), nil
		},
	}
	app := newTestApp(mock)

	_, err := app.findEeroID("12345", "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "eero not found") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestInspectEero(t *testing.T) {
	mock := &mockClient{
		GetEerosFn: func(networkID string) ([]api.Eero, error) {
			return testEeros(), nil
		},
		GetEeroRawFn: func(eeroID string) (json.RawMessage, error) {
			return json.RawMessage(`{"location":"Living Room","model":"eero Pro 6E"}`), nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.InspectEero("8318690"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "Living Room") {
		t.Error("output missing location in JSON")
	}
}

func TestRebootEero(t *testing.T) {
	var rebootedID string
	mock := &mockClient{
		GetEerosFn: func(networkID string) ([]api.Eero, error) {
			return testEeros(), nil
		},
		RebootEeroFn: func(eeroID string) error {
			rebootedID = eeroID
			return nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.RebootEero("8318690"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if rebootedID != "8318690" {
		t.Errorf("rebootedID = %q, want %q", rebootedID, "8318690")
	}
	if !strings.Contains(out, "Rebooting") {
		t.Error("output missing 'Rebooting'")
	}
	if !strings.Contains(out, "Living Room") {
		t.Error("output missing location")
	}
}

func TestEerosCommandRouting(t *testing.T) {
	mock := &mockClient{
		GetEerosFn: func(networkID string) ([]api.Eero, error) {
			return testEeros(), nil
		},
	}
	app := newTestApp(mock)

	// Test "list" routing
	captureStdout(t, func() {
		err := app.Eeros([]string{"list"})
		if err != nil {
			t.Fatalf("Eeros list routing: %v", err)
		}
	})

	// Test missing argument
	err := app.Eeros([]string{"inspect"})
	if err == nil || !strings.Contains(err.Error(), "usage") {
		t.Errorf("expected usage error, got: %v", err)
	}

	// Test unknown subcommand
	err = app.Eeros([]string{"invalid"})
	if err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Errorf("expected unknown error, got: %v", err)
	}
}
