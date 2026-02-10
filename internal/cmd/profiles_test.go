package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/dorin/eero-cli/internal/api"
)

func testProfiles() []api.Profile {
	return []api.Profile{
		{URL: "/2.2/networks/12345/profiles/prof1", Name: "Adults", Paused: false},
		{URL: "/2.2/networks/12345/profiles/prof2", Name: "Kids", Paused: true},
	}
}

func TestListProfiles(t *testing.T) {
	mock := &mockClient{
		GetProfilesFn: func(networkID string) ([]api.Profile, error) {
			return testProfiles(), nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.ListProfiles(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "Adults") {
		t.Error("output missing 'Adults'")
	}
	if !strings.Contains(out, "Kids") {
		t.Error("output missing 'Kids'")
	}
	if !strings.Contains(out, "paused") {
		t.Error("output missing 'paused' status")
	}
	if !strings.Contains(out, "Total: 2 profiles") {
		t.Errorf("output missing total count, got:\n%s", out)
	}
}

func TestListProfilesEmpty(t *testing.T) {
	mock := &mockClient{
		GetProfilesFn: func(networkID string) ([]api.Profile, error) {
			return []api.Profile{}, nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.ListProfiles(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "No profiles configured") {
		t.Errorf("expected empty message, got:\n%s", out)
	}
}

func TestFindProfileByID(t *testing.T) {
	mock := &mockClient{
		GetProfilesFn: func(networkID string) ([]api.Profile, error) {
			return testProfiles(), nil
		},
	}
	app := newTestApp(mock)

	id, err := app.findProfileID("12345", "prof1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "prof1" {
		t.Errorf("id = %q, want %q", id, "prof1")
	}
}

func TestFindProfileByName(t *testing.T) {
	mock := &mockClient{
		GetProfilesFn: func(networkID string) ([]api.Profile, error) {
			return testProfiles(), nil
		},
	}
	app := newTestApp(mock)

	id, err := app.findProfileID("12345", "Kids")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "prof2" {
		t.Errorf("id = %q, want %q", id, "prof2")
	}
}

func TestFindProfileByNameCaseInsensitive(t *testing.T) {
	mock := &mockClient{
		GetProfilesFn: func(networkID string) ([]api.Profile, error) {
			return testProfiles(), nil
		},
	}
	app := newTestApp(mock)

	id, err := app.findProfileID("12345", "adults")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "prof1" {
		t.Errorf("id = %q, want %q", id, "prof1")
	}
}

func TestFindProfileNotFound(t *testing.T) {
	mock := &mockClient{
		GetProfilesFn: func(networkID string) ([]api.Profile, error) {
			return testProfiles(), nil
		},
	}
	app := newTestApp(mock)

	_, err := app.findProfileID("12345", "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "profile not found") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestPauseProfile(t *testing.T) {
	var pausedID string
	var pauseValue bool
	mock := &mockClient{
		GetProfilesFn: func(networkID string) ([]api.Profile, error) {
			return testProfiles(), nil
		},
		PauseProfileFn: func(networkID, profileID string, pause bool) error {
			pausedID = profileID
			pauseValue = pause
			return nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.PauseProfile("prof1", true); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if pausedID != "prof1" {
		t.Errorf("pausedID = %q, want %q", pausedID, "prof1")
	}
	if !pauseValue {
		t.Error("pause = false, want true")
	}
	if !strings.Contains(out, "paused") {
		t.Error("output missing 'paused'")
	}
}

func TestInspectProfile(t *testing.T) {
	mock := &mockClient{
		GetProfilesFn: func(networkID string) ([]api.Profile, error) {
			return testProfiles(), nil
		},
		GetProfileRawFn: func(networkID, profileID string) (json.RawMessage, error) {
			return json.RawMessage(`{"name":"Adults","paused":false}`), nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.InspectProfile("prof1"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "Adults") {
		t.Error("output missing profile name in JSON")
	}
}

func TestAddDeviceToProfile(t *testing.T) {
	var gotDeviceURLs []string
	mock := &mockClient{
		GetProfilesFn: func(networkID string) ([]api.Profile, error) {
			return testProfiles(), nil
		},
		GetDevicesFn: func(networkID string) ([]api.Device, error) {
			return testDevices(), nil
		},
		GetProfileDetailsFn: func(networkID, profileID string) (*api.ProfileDetails, error) {
			return &api.ProfileDetails{
				URL:    "/2.2/networks/12345/profiles/prof1",
				Name:   "Adults",
				Paused: false,
				Devices: []struct {
					URL string `json:"url"`
				}{
					{URL: "/2.2/networks/12345/devices/aabbccdd1122"},
				},
			}, nil
		},
		SetProfileDevicesFn: func(networkID, profileID string, deviceURLs []string) error {
			gotDeviceURLs = deviceURLs
			return nil
		},
	}
	app := newTestApp(mock)

	// Add the "phone" device (eeff00112233) to Adults profile
	out := captureStdout(t, func() {
		if err := app.AddDeviceToProfile("prof1", "eeff00112233"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if len(gotDeviceURLs) != 2 {
		t.Fatalf("len(deviceURLs) = %d, want 2", len(gotDeviceURLs))
	}
	if !strings.Contains(out, "added") {
		t.Error("output missing 'added'")
	}
}

func TestAddDeviceToProfileAlreadyExists(t *testing.T) {
	mock := &mockClient{
		GetProfilesFn: func(networkID string) ([]api.Profile, error) {
			return testProfiles(), nil
		},
		GetDevicesFn: func(networkID string) ([]api.Device, error) {
			return testDevices(), nil
		},
		GetProfileDetailsFn: func(networkID, profileID string) (*api.ProfileDetails, error) {
			return &api.ProfileDetails{
				URL:  "/2.2/networks/12345/profiles/prof1",
				Name: "Adults",
				Devices: []struct {
					URL string `json:"url"`
				}{
					{URL: "/2.2/networks/12345/devices/aabbccdd1122"},
				},
			}, nil
		},
	}
	app := newTestApp(mock)

	err := app.AddDeviceToProfile("prof1", "aabbccdd1122")
	if err == nil {
		t.Fatal("expected error for duplicate device")
	}
	if !strings.Contains(err.Error(), "already in profile") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestRemoveDeviceFromProfile(t *testing.T) {
	var gotDeviceURLs []string
	mock := &mockClient{
		GetProfilesFn: func(networkID string) ([]api.Profile, error) {
			return testProfiles(), nil
		},
		GetDevicesFn: func(networkID string) ([]api.Device, error) {
			return testDevices(), nil
		},
		GetProfileDetailsFn: func(networkID, profileID string) (*api.ProfileDetails, error) {
			return &api.ProfileDetails{
				URL:  "/2.2/networks/12345/profiles/prof1",
				Name: "Adults",
				Devices: []struct {
					URL string `json:"url"`
				}{
					{URL: "/2.2/networks/12345/devices/aabbccdd1122"},
					{URL: "/2.2/networks/12345/devices/eeff00112233"},
				},
			}, nil
		},
		SetProfileDevicesFn: func(networkID, profileID string, deviceURLs []string) error {
			gotDeviceURLs = deviceURLs
			return nil
		},
	}
	app := newTestApp(mock)

	out := captureStdout(t, func() {
		if err := app.RemoveDeviceFromProfile("prof1", "aabbccdd1122"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if len(gotDeviceURLs) != 1 {
		t.Fatalf("len(deviceURLs) = %d, want 1", len(gotDeviceURLs))
	}
	if gotDeviceURLs[0] != "/2.2/networks/12345/devices/eeff00112233" {
		t.Errorf("remaining device = %q", gotDeviceURLs[0])
	}
	if !strings.Contains(out, "removed") {
		t.Error("output missing 'removed'")
	}
}

func TestRemoveDeviceNotInProfile(t *testing.T) {
	mock := &mockClient{
		GetProfilesFn: func(networkID string) ([]api.Profile, error) {
			return testProfiles(), nil
		},
		GetDevicesFn: func(networkID string) ([]api.Device, error) {
			return testDevices(), nil
		},
		GetProfileDetailsFn: func(networkID, profileID string) (*api.ProfileDetails, error) {
			return &api.ProfileDetails{
				URL:     "/2.2/networks/12345/profiles/prof1",
				Name:    "Adults",
				Devices: []struct{ URL string `json:"url"` }{},
			}, nil
		},
	}
	app := newTestApp(mock)

	err := app.RemoveDeviceFromProfile("prof1", "aabbccdd1122")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "not in profile") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestProfilesCommandRouting(t *testing.T) {
	mock := &mockClient{
		GetProfilesFn: func(networkID string) ([]api.Profile, error) {
			return testProfiles(), nil
		},
		PauseProfileFn: func(networkID, profileID string, pause bool) error {
			return nil
		},
	}
	app := newTestApp(mock)

	// Test "pause" routing
	captureStdout(t, func() {
		err := app.Profiles([]string{"pause", "prof1"})
		if err != nil {
			t.Fatalf("Profiles pause routing: %v", err)
		}
	})

	// Test missing argument
	err := app.Profiles([]string{"pause"})
	if err == nil || !strings.Contains(err.Error(), "usage") {
		t.Errorf("expected usage error, got: %v", err)
	}

	// Test unknown subcommand
	err = app.Profiles([]string{"invalid"})
	if err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Errorf("expected unknown error, got: %v", err)
	}
}

func TestPauseProfileAPIError(t *testing.T) {
	mock := &mockClient{
		GetProfilesFn: func(networkID string) ([]api.Profile, error) {
			return testProfiles(), nil
		},
		PauseProfileFn: func(networkID, profileID string, pause bool) error {
			return fmt.Errorf("API error: forbidden")
		},
	}
	app := newTestApp(mock)

	err := app.PauseProfile("prof1", true)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "forbidden") {
		t.Errorf("error = %q", err.Error())
	}
}
