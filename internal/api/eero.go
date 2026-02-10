package api

import "encoding/json"

// EeroAPI defines the interface for interacting with the Eero API.
// *Client satisfies this interface.
type EeroAPI interface {
	// Authentication
	Login(identity string) (*LoginResponse, error)
	LoginVerify(userToken, code string) error
	ValidateToken() bool
	SetToken(token string)

	// Account
	GetAccount() (*Account, error)

	// Devices
	GetDevices(networkID string) ([]Device, error)
	GetDeviceRaw(networkID, deviceID string) (json.RawMessage, error)
	UpdateDevice(networkID, deviceID string, updates map[string]interface{}) error
	PauseDevice(networkID, deviceID string, pause bool) error
	BlockDevice(networkID, deviceID string, block bool) error
	SetDeviceNickname(networkID, deviceID, nickname string) error

	// Profiles
	GetProfiles(networkID string) ([]Profile, error)
	GetProfileDetails(networkID, profileID string) (*ProfileDetails, error)
	GetProfileRaw(networkID, profileID string) (json.RawMessage, error)
	UpdateProfile(networkID, profileID string, updates map[string]interface{}) error
	SetProfileDevices(networkID, profileID string, deviceURLs []string) error
	PauseProfile(networkID, profileID string, pause bool) error

	// Eeros
	GetEeros(networkID string) ([]Eero, error)
	GetEeroRaw(eeroID string) (json.RawMessage, error)
	RebootEero(eeroID string) error

	// Guest Network
	GetGuestNetwork(networkID string) (*GuestNetwork, error)
	UpdateGuestNetwork(networkID string, updates map[string]interface{}) error
	EnableGuestNetwork(networkID string, enable bool) error
	SetGuestNetworkPassword(networkID, password string) error

	// Network
	Reboot(networkID string) error

	// Reservations
	GetReservations(networkID string) ([]Reservation, error)
	GetReservationRaw(networkID, reservationID string) (json.RawMessage, error)
	CreateReservation(networkID, ip, mac, description string) error
	DeleteReservation(networkID, reservationID string) error
}
