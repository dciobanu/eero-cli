// Package api provides the Eero API client
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	baseURL   = "https://api-user.e2ro.com"
	userAgent = "eero-ios/2.16.0 (iPhone8,1; iOS 11.3)"
)

// Client is the Eero API client
type Client struct {
	token      string
	httpClient *http.Client
}

// New creates a new Eero API client
func New(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToken updates the client's authentication token
func (c *Client) SetToken(token string) {
	c.token = token
}

// request makes an HTTP request to the Eero API
func (c *Client) request(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Cookie", "s="+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr APIError
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Meta.Error != "" {
			return nil, fmt.Errorf("API error: %s", apiErr.Meta.Error)
		}
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// APIError represents an error response from the Eero API
type APIError struct {
	Meta struct {
		Code  int    `json:"code"`
		Error string `json:"error"`
	} `json:"meta"`
}

// APIResponse wraps the standard API response format
type APIResponse struct {
	Meta struct {
		Code      int    `json:"code"`
		ServerID  string `json:"server_id"`
		Timestamp int64  `json:"timestamp"`
	} `json:"meta"`
	Data json.RawMessage `json:"data"`
}

// LoginResponse contains the response from the login endpoint
type LoginResponse struct {
	UserToken string `json:"user_token"`
}

// Login initiates the authentication flow
func (c *Client) Login(identity string) (*LoginResponse, error) {
	payload := map[string]string{"login": identity}
	data, err := c.request("POST", "/2.2/login", payload)
	if err != nil {
		return nil, err
	}

	var resp APIResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(resp.Data, &loginResp); err != nil {
		return nil, fmt.Errorf("parsing login data: %w", err)
	}

	return &loginResp, nil
}

// LoginVerify completes the authentication with a verification code
func (c *Client) LoginVerify(userToken, code string) error {
	c.SetToken(userToken)
	payload := map[string]string{"code": code}
	_, err := c.request("POST", "/2.2/login/verify", payload)
	return err
}

// Account represents the user account
type Account struct {
	Email    string    `json:"email_verified"`
	Phone    string    `json:"phone_verified"`
	Name     string    `json:"name"`
	Networks []Network `json:"networks"`
}

// Network represents an Eero network
type Network struct {
	URL     string `json:"url"`
	Name    string `json:"name"`
	Premium bool   `json:"premium_status"`
}

// GetAccount returns the current account information
func (c *Client) GetAccount() (*Account, error) {
	data, err := c.request("GET", "/2.2/account", nil)
	if err != nil {
		return nil, err
	}

	var resp APIResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	var account Account
	if err := json.Unmarshal(resp.Data, &account); err != nil {
		return nil, fmt.Errorf("parsing account data: %w", err)
	}

	return &account, nil
}

// Device represents a connected device
type Device struct {
	URL       string `json:"url"`
	MAC       string `json:"mac"`
	Hostname  string `json:"hostname"`
	Nickname  string `json:"nickname"`
	IP        string `json:"ip"`
	Connected bool   `json:"connected"`
	Wireless  bool   `json:"wireless"`
	Paused    bool   `json:"paused"`
	Blocked   bool   `json:"blocked"`
	Profile   *struct {
		URL  string `json:"url"`
		Name string `json:"name"`
	} `json:"profile"`
	ConnectionType string `json:"connection_type"`
	DeviceType     string `json:"device_type"`
}

// DisplayName returns the best available name for the device
func (d *Device) DisplayName() string {
	if d.Nickname != "" {
		return d.Nickname
	}
	if d.Hostname != "" {
		return d.Hostname
	}
	return d.MAC
}

// GetDevices returns all devices on the network
func (c *Client) GetDevices(networkID string) ([]Device, error) {
	path := fmt.Sprintf("/2.2/networks/%s/devices", networkID)
	data, err := c.request("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp APIResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	var devices []Device
	if err := json.Unmarshal(resp.Data, &devices); err != nil {
		return nil, fmt.Errorf("parsing devices data: %w", err)
	}

	return devices, nil
}

// UpdateDevice modifies a device's settings
func (c *Client) UpdateDevice(networkID, deviceID string, updates map[string]interface{}) error {
	path := fmt.Sprintf("/2.2/networks/%s/devices/%s", networkID, deviceID)
	_, err := c.request("PUT", path, updates)
	return err
}

// PauseDevice pauses or unpauses a device
func (c *Client) PauseDevice(networkID, deviceID string, pause bool) error {
	return c.UpdateDevice(networkID, deviceID, map[string]interface{}{"paused": pause})
}

// BlockDevice blocks or unblocks a device
func (c *Client) BlockDevice(networkID, deviceID string, block bool) error {
	return c.UpdateDevice(networkID, deviceID, map[string]interface{}{"blocked": block})
}

// SetDeviceNickname sets a device's nickname
func (c *Client) SetDeviceNickname(networkID, deviceID, nickname string) error {
	return c.UpdateDevice(networkID, deviceID, map[string]interface{}{"nickname": nickname})
}

// Profile represents a family profile
type Profile struct {
	URL    string `json:"url"`
	Name   string `json:"name"`
	Paused bool   `json:"paused"`
}

// GetProfiles returns all profiles on the network
func (c *Client) GetProfiles(networkID string) ([]Profile, error) {
	path := fmt.Sprintf("/2.2/networks/%s/profiles", networkID)
	data, err := c.request("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp APIResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	var profiles []Profile
	if err := json.Unmarshal(resp.Data, &profiles); err != nil {
		return nil, fmt.Errorf("parsing profiles data: %w", err)
	}

	return profiles, nil
}

// UpdateProfile modifies a profile's settings
func (c *Client) UpdateProfile(networkID, profileID string, updates map[string]interface{}) error {
	path := fmt.Sprintf("/2.2/networks/%s/profiles/%s", networkID, profileID)
	_, err := c.request("PUT", path, updates)
	return err
}

// PauseProfile pauses or unpauses a profile
func (c *Client) PauseProfile(networkID, profileID string, pause bool) error {
	return c.UpdateProfile(networkID, profileID, map[string]interface{}{"paused": pause})
}

// GuestNetwork represents guest network settings
type GuestNetwork struct {
	Enabled  bool   `json:"enabled"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// GetGuestNetwork returns the guest network settings
func (c *Client) GetGuestNetwork(networkID string) (*GuestNetwork, error) {
	path := fmt.Sprintf("/2.2/networks/%s/guestnetwork", networkID)
	data, err := c.request("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp APIResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	var gn GuestNetwork
	if err := json.Unmarshal(resp.Data, &gn); err != nil {
		return nil, fmt.Errorf("parsing guest network data: %w", err)
	}

	return &gn, nil
}

// UpdateGuestNetwork modifies the guest network settings
func (c *Client) UpdateGuestNetwork(networkID string, updates map[string]interface{}) error {
	path := fmt.Sprintf("/2.2/networks/%s/guestnetwork", networkID)
	_, err := c.request("PUT", path, updates)
	return err
}

// EnableGuestNetwork enables or disables the guest network
func (c *Client) EnableGuestNetwork(networkID string, enable bool) error {
	return c.UpdateGuestNetwork(networkID, map[string]interface{}{"enabled": enable})
}

// SetGuestNetworkPassword sets the guest network password
func (c *Client) SetGuestNetworkPassword(networkID, password string) error {
	return c.UpdateGuestNetwork(networkID, map[string]interface{}{"password": password})
}

// Reboot reboots the entire network
func (c *Client) Reboot(networkID string) error {
	path := fmt.Sprintf("/2.2/networks/%s/reboot", networkID)
	_, err := c.request("POST", path, nil)
	return err
}

// ValidateToken checks if the current token is valid
func (c *Client) ValidateToken() bool {
	if c.token == "" {
		return false
	}
	_, err := c.GetAccount()
	return err == nil
}

// ExtractNetworkID extracts the network ID from a URL path like "/2.2/networks/12345"
func ExtractNetworkID(url string) string {
	// URL format: /2.2/networks/{id}
	if len(url) > 14 { // len("/2.2/networks/") = 14
		return url[14:]
	}
	return url
}

// ExtractDeviceID extracts the device ID from a URL path
func ExtractDeviceID(url string) string {
	// URL format: /2.2/devices/{id}
	if len(url) > 12 { // len("/2.2/devices/") = 13
		return url[13:]
	}
	return url
}

// ExtractProfileID extracts the profile ID from a URL path
func ExtractProfileID(url string) string {
	// URL format: /2.2/profiles/{id}
	if len(url) > 13 { // len("/2.2/profiles/") = 14
		return url[14:]
	}
	return url
}
