package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/dorin/eero-cli/internal/api"
	"github.com/dorin/eero-cli/internal/config"
)

// mockClient implements api.EeroAPI with function fields for testing.
// Each method checks for a corresponding function field; if nil, it panics
// to surface unexpected calls during tests.
type mockClient struct {
	LoginFn                 func(identity string) (*api.LoginResponse, error)
	LoginVerifyFn           func(userToken, code string) error
	ValidateTokenFn         func() bool
	SetTokenFn              func(token string)
	GetAccountFn            func() (*api.Account, error)
	GetDevicesFn            func(networkID string) ([]api.Device, error)
	GetDeviceRawFn          func(networkID, deviceID string) (json.RawMessage, error)
	UpdateDeviceFn          func(networkID, deviceID string, updates map[string]interface{}) error
	PauseDeviceFn           func(networkID, deviceID string, pause bool) error
	BlockDeviceFn           func(networkID, deviceID string, block bool) error
	SetDeviceNicknameFn     func(networkID, deviceID, nickname string) error
	GetProfilesFn           func(networkID string) ([]api.Profile, error)
	GetProfileDetailsFn     func(networkID, profileID string) (*api.ProfileDetails, error)
	GetProfileRawFn         func(networkID, profileID string) (json.RawMessage, error)
	UpdateProfileFn         func(networkID, profileID string, updates map[string]interface{}) error
	SetProfileDevicesFn     func(networkID, profileID string, deviceURLs []string) error
	PauseProfileFn          func(networkID, profileID string, pause bool) error
	GetEerosFn              func(networkID string) ([]api.Eero, error)
	GetEeroRawFn            func(eeroID string) (json.RawMessage, error)
	RebootEeroFn            func(eeroID string) error
	GetGuestNetworkFn       func(networkID string) (*api.GuestNetwork, error)
	UpdateGuestNetworkFn    func(networkID string, updates map[string]interface{}) error
	EnableGuestNetworkFn    func(networkID string, enable bool) error
	SetGuestNetworkPasswordFn func(networkID, password string) error
	RebootFn                func(networkID string) error
	GetReservationsFn       func(networkID string) ([]api.Reservation, error)
	GetReservationRawFn     func(networkID, reservationID string) (json.RawMessage, error)
	CreateReservationFn     func(networkID, ip, mac, description string) error
	DeleteReservationFn     func(networkID, reservationID string) error
}

func (m *mockClient) Login(identity string) (*api.LoginResponse, error) {
	if m.LoginFn != nil {
		return m.LoginFn(identity)
	}
	panic("mockClient.Login not set")
}

func (m *mockClient) LoginVerify(userToken, code string) error {
	if m.LoginVerifyFn != nil {
		return m.LoginVerifyFn(userToken, code)
	}
	panic("mockClient.LoginVerify not set")
}

func (m *mockClient) ValidateToken() bool {
	if m.ValidateTokenFn != nil {
		return m.ValidateTokenFn()
	}
	return true
}

func (m *mockClient) SetToken(token string) {
	if m.SetTokenFn != nil {
		m.SetTokenFn(token)
	}
}

func (m *mockClient) GetAccount() (*api.Account, error) {
	if m.GetAccountFn != nil {
		return m.GetAccountFn()
	}
	panic("mockClient.GetAccount not set")
}

func (m *mockClient) GetDevices(networkID string) ([]api.Device, error) {
	if m.GetDevicesFn != nil {
		return m.GetDevicesFn(networkID)
	}
	panic("mockClient.GetDevices not set")
}

func (m *mockClient) GetDeviceRaw(networkID, deviceID string) (json.RawMessage, error) {
	if m.GetDeviceRawFn != nil {
		return m.GetDeviceRawFn(networkID, deviceID)
	}
	panic("mockClient.GetDeviceRaw not set")
}

func (m *mockClient) UpdateDevice(networkID, deviceID string, updates map[string]interface{}) error {
	if m.UpdateDeviceFn != nil {
		return m.UpdateDeviceFn(networkID, deviceID, updates)
	}
	panic("mockClient.UpdateDevice not set")
}

func (m *mockClient) PauseDevice(networkID, deviceID string, pause bool) error {
	if m.PauseDeviceFn != nil {
		return m.PauseDeviceFn(networkID, deviceID, pause)
	}
	panic("mockClient.PauseDevice not set")
}

func (m *mockClient) BlockDevice(networkID, deviceID string, block bool) error {
	if m.BlockDeviceFn != nil {
		return m.BlockDeviceFn(networkID, deviceID, block)
	}
	panic("mockClient.BlockDevice not set")
}

func (m *mockClient) SetDeviceNickname(networkID, deviceID, nickname string) error {
	if m.SetDeviceNicknameFn != nil {
		return m.SetDeviceNicknameFn(networkID, deviceID, nickname)
	}
	panic("mockClient.SetDeviceNickname not set")
}

func (m *mockClient) GetProfiles(networkID string) ([]api.Profile, error) {
	if m.GetProfilesFn != nil {
		return m.GetProfilesFn(networkID)
	}
	panic("mockClient.GetProfiles not set")
}

func (m *mockClient) GetProfileDetails(networkID, profileID string) (*api.ProfileDetails, error) {
	if m.GetProfileDetailsFn != nil {
		return m.GetProfileDetailsFn(networkID, profileID)
	}
	panic("mockClient.GetProfileDetails not set")
}

func (m *mockClient) GetProfileRaw(networkID, profileID string) (json.RawMessage, error) {
	if m.GetProfileRawFn != nil {
		return m.GetProfileRawFn(networkID, profileID)
	}
	panic("mockClient.GetProfileRaw not set")
}

func (m *mockClient) UpdateProfile(networkID, profileID string, updates map[string]interface{}) error {
	if m.UpdateProfileFn != nil {
		return m.UpdateProfileFn(networkID, profileID, updates)
	}
	panic("mockClient.UpdateProfile not set")
}

func (m *mockClient) SetProfileDevices(networkID, profileID string, deviceURLs []string) error {
	if m.SetProfileDevicesFn != nil {
		return m.SetProfileDevicesFn(networkID, profileID, deviceURLs)
	}
	panic("mockClient.SetProfileDevices not set")
}

func (m *mockClient) PauseProfile(networkID, profileID string, pause bool) error {
	if m.PauseProfileFn != nil {
		return m.PauseProfileFn(networkID, profileID, pause)
	}
	panic("mockClient.PauseProfile not set")
}

func (m *mockClient) GetEeros(networkID string) ([]api.Eero, error) {
	if m.GetEerosFn != nil {
		return m.GetEerosFn(networkID)
	}
	panic("mockClient.GetEeros not set")
}

func (m *mockClient) GetEeroRaw(eeroID string) (json.RawMessage, error) {
	if m.GetEeroRawFn != nil {
		return m.GetEeroRawFn(eeroID)
	}
	panic("mockClient.GetEeroRaw not set")
}

func (m *mockClient) RebootEero(eeroID string) error {
	if m.RebootEeroFn != nil {
		return m.RebootEeroFn(eeroID)
	}
	panic("mockClient.RebootEero not set")
}

func (m *mockClient) GetGuestNetwork(networkID string) (*api.GuestNetwork, error) {
	if m.GetGuestNetworkFn != nil {
		return m.GetGuestNetworkFn(networkID)
	}
	panic("mockClient.GetGuestNetwork not set")
}

func (m *mockClient) UpdateGuestNetwork(networkID string, updates map[string]interface{}) error {
	if m.UpdateGuestNetworkFn != nil {
		return m.UpdateGuestNetworkFn(networkID, updates)
	}
	panic("mockClient.UpdateGuestNetwork not set")
}

func (m *mockClient) EnableGuestNetwork(networkID string, enable bool) error {
	if m.EnableGuestNetworkFn != nil {
		return m.EnableGuestNetworkFn(networkID, enable)
	}
	panic("mockClient.EnableGuestNetwork not set")
}

func (m *mockClient) SetGuestNetworkPassword(networkID, password string) error {
	if m.SetGuestNetworkPasswordFn != nil {
		return m.SetGuestNetworkPasswordFn(networkID, password)
	}
	panic("mockClient.SetGuestNetworkPassword not set")
}

func (m *mockClient) Reboot(networkID string) error {
	if m.RebootFn != nil {
		return m.RebootFn(networkID)
	}
	panic("mockClient.Reboot not set")
}

func (m *mockClient) GetReservations(networkID string) ([]api.Reservation, error) {
	if m.GetReservationsFn != nil {
		return m.GetReservationsFn(networkID)
	}
	panic("mockClient.GetReservations not set")
}

func (m *mockClient) GetReservationRaw(networkID, reservationID string) (json.RawMessage, error) {
	if m.GetReservationRawFn != nil {
		return m.GetReservationRawFn(networkID, reservationID)
	}
	panic("mockClient.GetReservationRaw not set")
}

func (m *mockClient) CreateReservation(networkID, ip, mac, description string) error {
	if m.CreateReservationFn != nil {
		return m.CreateReservationFn(networkID, ip, mac, description)
	}
	panic("mockClient.CreateReservation not set")
}

func (m *mockClient) DeleteReservation(networkID, reservationID string) error {
	if m.DeleteReservationFn != nil {
		return m.DeleteReservationFn(networkID, reservationID)
	}
	panic("mockClient.DeleteReservation not set")
}

// newTestApp creates an App with the given mock client and a pre-configured
// network ID, bypassing EnsureAuth / EnsureNetwork.
func newTestApp(mock *mockClient) *App {
	return &App{
		Config: &config.Config{
			Token:     "test-token",
			NetworkID: "12345",
		},
		Client: mock,
	}
}

// captureStdout redirects os.Stdout for the duration of fn and returns
// whatever was written. This avoids refactoring all commands to accept
// an io.Writer.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	// Run the function
	fn()

	w.Close()
	os.Stdout = old

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("reading captured stdout: %v", err)
	}
	return string(out)
}

// testDevices returns a standard set of devices for testing
func testDevices() []api.Device {
	return []api.Device{
		{
			URL:      "/2.2/networks/12345/devices/aabbccdd1122",
			MAC:      "AA:BB:CC:DD:11:22",
			Hostname: "laptop",
			Nickname: "My Laptop",
			IP:       "192.168.1.100",
			Connected: true,
			Wireless:  true,
			Profile: &struct {
				URL  string `json:"url"`
				Name string `json:"name"`
			}{URL: "/2.2/networks/12345/profiles/prof1", Name: "Adults"},
		},
		{
			URL:       "/2.2/networks/12345/devices/eeff00112233",
			MAC:       "EE:FF:00:11:22:33",
			Hostname:  "phone",
			IP:        "192.168.1.101",
			Connected: false,
			Wireless:  true,
			IsPrivate: true,
		},
		{
			URL:       "/2.2/networks/12345/devices/112233445566",
			MAC:       "11:22:33:44:55:66",
			Hostname:  "server",
			Nickname:  "NAS",
			IP:        "192.168.1.10",
			Connected: true,
			Wireless:  false,
		},
	}
}

// errNotFound is a convenience error for tests
var errNotFound = fmt.Errorf("not found")
