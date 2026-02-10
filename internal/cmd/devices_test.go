package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/dorin/eero-cli/internal/api"
)

func TestFindDeviceByExactID(t *testing.T) {
	mock := &mockClient{
		GetDevicesFn: func(networkID string) ([]api.Device, error) {
			return testDevices(), nil
		},
	}
	app := newTestApp(mock)

	id, err := app.findDeviceID("12345", "aabbccdd1122")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "aabbccdd1122" {
		t.Errorf("id = %q, want %q", id, "aabbccdd1122")
	}
}

func TestFindDeviceByPartialID(t *testing.T) {
	mock := &mockClient{
		GetDevicesFn: func(networkID string) ([]api.Device, error) {
			return testDevices(), nil
		},
	}
	app := newTestApp(mock)

	id, err := app.findDeviceID("12345", "aabb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "aabbccdd1122" {
		t.Errorf("id = %q, want %q", id, "aabbccdd1122")
	}
}

func TestFindDeviceByMAC(t *testing.T) {
	mock := &mockClient{
		GetDevicesFn: func(networkID string) ([]api.Device, error) {
			return testDevices(), nil
		},
	}
	app := newTestApp(mock)

	id, err := app.findDeviceID("12345", "AA:BB:CC:DD:11:22")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "aabbccdd1122" {
		t.Errorf("id = %q, want %q", id, "aabbccdd1122")
	}
}

func TestFindDeviceByMACWithoutColons(t *testing.T) {
	mock := &mockClient{
		GetDevicesFn: func(networkID string) ([]api.Device, error) {
			return testDevices(), nil
		},
	}
	app := newTestApp(mock)

	id, err := app.findDeviceID("12345", "aabbccdd1122")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "aabbccdd1122" {
		t.Errorf("id = %q, want %q", id, "aabbccdd1122")
	}
}

func TestFindDeviceByName(t *testing.T) {
	mock := &mockClient{
		GetDevicesFn: func(networkID string) ([]api.Device, error) {
			return testDevices(), nil
		},
	}
	app := newTestApp(mock)

	id, err := app.findDeviceID("12345", "My Laptop")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "aabbccdd1122" {
		t.Errorf("id = %q, want %q", id, "aabbccdd1122")
	}
}

func TestFindDeviceByHostname(t *testing.T) {
	mock := &mockClient{
		GetDevicesFn: func(networkID string) ([]api.Device, error) {
			return testDevices(), nil
		},
	}
	app := newTestApp(mock)

	// "phone" has no nickname, so DisplayName() returns hostname
	id, err := app.findDeviceID("12345", "phone")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "eeff00112233" {
		t.Errorf("id = %q, want %q", id, "eeff00112233")
	}
}

func TestFindDeviceNotFound(t *testing.T) {
	mock := &mockClient{
		GetDevicesFn: func(networkID string) ([]api.Device, error) {
			return testDevices(), nil
		},
	}
	app := newTestApp(mock)

	_, err := app.findDeviceID("12345", "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "device not found") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestListDevicesNoFilter(t *testing.T) {
	mock := &mockClient{
		GetDevicesFn: func(networkID string) ([]api.Device, error) {
			return testDevices(), nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.ListDevices(DeviceFilters{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "My Laptop") {
		t.Error("output missing 'My Laptop'")
	}
	if !strings.Contains(out, "NAS") {
		t.Error("output missing 'NAS'")
	}
	if !strings.Contains(out, "Total: 3 devices") {
		t.Errorf("output missing total count, got:\n%s", out)
	}
}

func TestListDevicesWiredFilter(t *testing.T) {
	mock := &mockClient{
		GetDevicesFn: func(networkID string) ([]api.Device, error) {
			return testDevices(), nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.ListDevices(DeviceFilters{Wired: true}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// Only the NAS (wired) should appear
	if !strings.Contains(out, "NAS") {
		t.Error("output missing wired device 'NAS'")
	}
	if strings.Contains(out, "My Laptop") {
		t.Error("output should not contain wireless device 'My Laptop'")
	}
	if !strings.Contains(out, "1 devices") {
		t.Errorf("expected 1 filtered device, got:\n%s", out)
	}
}

func TestListDevicesOnlineFilter(t *testing.T) {
	mock := &mockClient{
		GetDevicesFn: func(networkID string) ([]api.Device, error) {
			return testDevices(), nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.ListDevices(DeviceFilters{Online: true}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// Phone is offline, should be filtered out
	if strings.Contains(out, "phone") {
		t.Error("output should not contain offline device 'phone'")
	}
	if !strings.Contains(out, "2 devices") {
		t.Errorf("expected 2 online devices, got:\n%s", out)
	}
}

func TestListDevicesPrivateFilter(t *testing.T) {
	mock := &mockClient{
		GetDevicesFn: func(networkID string) ([]api.Device, error) {
			return testDevices(), nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.ListDevices(DeviceFilters{Private: true}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// Only phone is private
	if !strings.Contains(out, "phone") {
		t.Error("output missing private device 'phone'")
	}
	if strings.Contains(out, "My Laptop") {
		t.Error("output should not contain non-private device")
	}
}

func TestListDevicesProfileFilter(t *testing.T) {
	mock := &mockClient{
		GetDevicesFn: func(networkID string) ([]api.Device, error) {
			return testDevices(), nil
		},
		GetProfilesFn: func(networkID string) ([]api.Profile, error) {
			return []api.Profile{
				{URL: "/2.2/networks/12345/profiles/prof1", Name: "Adults"},
			}, nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.ListDevices(DeviceFilters{Profile: "Adults"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "My Laptop") {
		t.Error("output missing device in Adults profile")
	}
	if strings.Contains(out, "phone") {
		t.Error("output should not contain device without Adults profile")
	}
}

func TestPauseDevice(t *testing.T) {
	var pausedID string
	var pauseValue bool
	mock := &mockClient{
		GetDevicesFn: func(networkID string) ([]api.Device, error) {
			return testDevices(), nil
		},
		PauseDeviceFn: func(networkID, deviceID string, pause bool) error {
			pausedID = deviceID
			pauseValue = pause
			return nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.PauseDevice("aabbccdd1122", true); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if pausedID != "aabbccdd1122" {
		t.Errorf("pausedID = %q, want %q", pausedID, "aabbccdd1122")
	}
	if !pauseValue {
		t.Error("pause = false, want true")
	}
	if !strings.Contains(out, "paused") {
		t.Error("output missing 'paused'")
	}
}

func TestUnpauseDevice(t *testing.T) {
	var pauseValue bool
	mock := &mockClient{
		GetDevicesFn: func(networkID string) ([]api.Device, error) {
			return testDevices(), nil
		},
		PauseDeviceFn: func(networkID, deviceID string, pause bool) error {
			pauseValue = pause
			return nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.PauseDevice("aabbccdd1122", false); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if pauseValue {
		t.Error("pause = true, want false")
	}
	if !strings.Contains(out, "unpaused") {
		t.Error("output missing 'unpaused'")
	}
}

func TestBlockDevice(t *testing.T) {
	var blockedID string
	mock := &mockClient{
		GetDevicesFn: func(networkID string) ([]api.Device, error) {
			return testDevices(), nil
		},
		BlockDeviceFn: func(networkID, deviceID string, block bool) error {
			blockedID = deviceID
			return nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.BlockDevice("aabbccdd1122", true); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if blockedID != "aabbccdd1122" {
		t.Errorf("blockedID = %q", blockedID)
	}
	if !strings.Contains(out, "blocked") {
		t.Error("output missing 'blocked'")
	}
}

func TestRenameDevice(t *testing.T) {
	var gotNickname string
	mock := &mockClient{
		GetDevicesFn: func(networkID string) ([]api.Device, error) {
			return testDevices(), nil
		},
		SetDeviceNicknameFn: func(networkID, deviceID, nickname string) error {
			gotNickname = nickname
			return nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.RenameDevice("aabbccdd1122", "New Name"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if gotNickname != "New Name" {
		t.Errorf("nickname = %q, want %q", gotNickname, "New Name")
	}
	if !strings.Contains(out, "renamed") {
		t.Error("output missing 'renamed'")
	}
}

func TestInspectDevice(t *testing.T) {
	mock := &mockClient{
		GetDevicesFn: func(networkID string) ([]api.Device, error) {
			return testDevices(), nil
		},
		GetDeviceRawFn: func(networkID, deviceID string) (json.RawMessage, error) {
			return json.RawMessage(`{"mac":"AA:BB:CC:DD:11:22","nickname":"My Laptop"}`), nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.InspectDevice("aabbccdd1122"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "My Laptop") {
		t.Error("output missing device nickname in JSON")
	}
}

func TestPauseDeviceAPIError(t *testing.T) {
	mock := &mockClient{
		GetDevicesFn: func(networkID string) ([]api.Device, error) {
			return testDevices(), nil
		},
		PauseDeviceFn: func(networkID, deviceID string, pause bool) error {
			return fmt.Errorf("API error: forbidden")
		},
	}
	app := newTestApp(mock)

	err := app.PauseDevice("aabbccdd1122", true)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "forbidden") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestDevicesCommandRouting(t *testing.T) {
	mock := &mockClient{
		GetDevicesFn: func(networkID string) ([]api.Device, error) {
			return testDevices(), nil
		},
		PauseDeviceFn: func(networkID, deviceID string, pause bool) error {
			return nil
		},
	}
	app := newTestApp(mock)

	// Test "pause" subcommand routing
	captureStdout(t, func() {
		err := app.Devices([]string{"pause", "aabbccdd1122"})
		if err != nil {
			t.Fatalf("Devices pause routing: %v", err)
		}
	})

	// Test missing argument
	err := app.Devices([]string{"pause"})
	if err == nil || !strings.Contains(err.Error(), "usage") {
		t.Errorf("expected usage error, got: %v", err)
	}

	// Test unknown subcommand
	err = app.Devices([]string{"invalid"})
	if err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Errorf("expected unknown error, got: %v", err)
	}
}
