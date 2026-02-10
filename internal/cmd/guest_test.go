package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dorin/eero-cli/internal/api"
)

func TestGuestStatusEnabled(t *testing.T) {
	mock := &mockClient{
		GetGuestNetworkFn: func(networkID string) (*api.GuestNetwork, error) {
			return &api.GuestNetwork{
				Enabled:  true,
				Name:     "Home Guest",
				Password: "guestpass123",
			}, nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.GuestStatus(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "enabled") {
		t.Error("output missing 'enabled'")
	}
	if !strings.Contains(out, "Home Guest") {
		t.Error("output missing network name")
	}
	if !strings.Contains(out, "guestpass123") {
		t.Error("output missing password")
	}
}

func TestGuestStatusDisabled(t *testing.T) {
	mock := &mockClient{
		GetGuestNetworkFn: func(networkID string) (*api.GuestNetwork, error) {
			return &api.GuestNetwork{
				Enabled:  false,
				Name:     "Home Guest",
				Password: "guestpass123",
			}, nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.GuestStatus(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "disabled") {
		t.Error("output missing 'disabled'")
	}
	// Password should not be shown when disabled
	if strings.Contains(out, "guestpass123") {
		t.Error("password should not be shown when guest network is disabled")
	}
}

func TestGuestEnable(t *testing.T) {
	var enableValue bool
	mock := &mockClient{
		EnableGuestNetworkFn: func(networkID string, enable bool) error {
			enableValue = enable
			return nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.GuestEnable(true); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !enableValue {
		t.Error("enable = false, want true")
	}
	if !strings.Contains(out, "enabled") {
		t.Error("output missing 'enabled'")
	}
}

func TestGuestDisable(t *testing.T) {
	var enableValue bool
	mock := &mockClient{
		EnableGuestNetworkFn: func(networkID string, enable bool) error {
			enableValue = enable
			return nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.GuestEnable(false); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if enableValue {
		t.Error("enable = true, want false")
	}
	if !strings.Contains(out, "disabled") {
		t.Error("output missing 'disabled'")
	}
}

func TestGuestPassword(t *testing.T) {
	var gotPassword string
	mock := &mockClient{
		SetGuestNetworkPasswordFn: func(networkID, password string) error {
			gotPassword = password
			return nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.GuestPassword("newpass123"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if gotPassword != "newpass123" {
		t.Errorf("password = %q, want %q", gotPassword, "newpass123")
	}
	if !strings.Contains(out, "password has been updated") {
		t.Error("output missing confirmation message")
	}
}

func TestGuestPasswordError(t *testing.T) {
	mock := &mockClient{
		SetGuestNetworkPasswordFn: func(networkID, password string) error {
			return fmt.Errorf("API error: bad request")
		},
	}
	app := newTestApp(mock)

	err := app.GuestPassword("short")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "bad request") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestGuestCommandRouting(t *testing.T) {
	mock := &mockClient{
		EnableGuestNetworkFn: func(networkID string, enable bool) error {
			return nil
		},
	}
	app := newTestApp(mock)

	// Test "enable" routing
	captureStdout(t, func() {
		err := app.Guest([]string{"enable"})
		if err != nil {
			t.Fatalf("Guest enable routing: %v", err)
		}
	})

	// Test "disable" routing
	captureStdout(t, func() {
		err := app.Guest([]string{"disable"})
		if err != nil {
			t.Fatalf("Guest disable routing: %v", err)
		}
	})

	// Test missing password argument
	err := app.Guest([]string{"password"})
	if err == nil || !strings.Contains(err.Error(), "usage") {
		t.Errorf("expected usage error, got: %v", err)
	}

	// Test unknown subcommand
	err = app.Guest([]string{"invalid"})
	if err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Errorf("expected unknown error, got: %v", err)
	}
}
