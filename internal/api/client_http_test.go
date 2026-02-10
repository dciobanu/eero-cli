package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// loadFixture reads a JSON fixture from testdata/
func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("loading fixture %s: %v", name, err)
	}
	return data
}

// newTestServer creates an httptest server and a client pointed at it.
// The handler receives every request; the caller decides what to return.
func newTestServer(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client := New("test-token")
	client.SetBaseURL(srv.URL)
	return client, srv
}

// --- Auth cookie ---

func TestRequestSetsAuthCookie(t *testing.T) {
	var gotCookie string
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotCookie = r.Header.Get("Cookie")
		w.Write(loadFixture(t, "account.json"))
	})

	_, err := client.GetAccount()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotCookie != "s=test-token" {
		t.Errorf("Cookie = %q, want %q", gotCookie, "s=test-token")
	}
}

func TestRequestSetsUserAgent(t *testing.T) {
	var gotUA string
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.Write(loadFixture(t, "account.json"))
	})

	client.GetAccount()
	if gotUA != userAgent {
		t.Errorf("User-Agent = %q, want %q", gotUA, userAgent)
	}
}

// --- Account ---

func TestGetAccount(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/2.2/account" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Write(loadFixture(t, "account.json"))
	})

	account, err := client.GetAccount()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if account.Name != "Test User" {
		t.Errorf("Name = %q, want %q", account.Name, "Test User")
	}
	if account.Email.Value != "test@example.com" {
		t.Errorf("Email = %q, want %q", account.Email.Value, "test@example.com")
	}
	if account.Networks.Count != 1 {
		t.Errorf("Networks.Count = %d, want 1", account.Networks.Count)
	}
	if len(account.Networks.Data) != 1 {
		t.Fatalf("len(Networks.Data) = %d, want 1", len(account.Networks.Data))
	}
	if account.Networks.Data[0].Name != "Home Network" {
		t.Errorf("Network.Name = %q, want %q", account.Networks.Data[0].Name, "Home Network")
	}
}

// --- Devices ---

func TestGetDevices(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/2.2/networks/12345/devices" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Write(loadFixture(t, "devices.json"))
	})

	devices, err := client.GetDevices("12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 3 {
		t.Fatalf("len(devices) = %d, want 3", len(devices))
	}

	d := devices[0]
	if d.Nickname != "My Laptop" {
		t.Errorf("Nickname = %q, want %q", d.Nickname, "My Laptop")
	}
	if d.MAC != "AA:BB:CC:DD:11:22" {
		t.Errorf("MAC = %q, want %q", d.MAC, "AA:BB:CC:DD:11:22")
	}
	if d.IP != "192.168.1.100" {
		t.Errorf("IP = %q, want %q", d.IP, "192.168.1.100")
	}
	if !d.Connected {
		t.Error("Connected = false, want true")
	}
	if !d.Wireless {
		t.Error("Wireless = false, want true")
	}
	if d.Profile == nil {
		t.Fatal("Profile = nil, want non-nil")
	}
	if d.Profile.Name != "Adults" {
		t.Errorf("Profile.Name = %q, want %q", d.Profile.Name, "Adults")
	}

	// second device: no nickname, no profile, private
	d2 := devices[1]
	if d2.Nickname != "" {
		t.Errorf("d2.Nickname = %q, want empty", d2.Nickname)
	}
	if d2.Profile != nil {
		t.Errorf("d2.Profile = %v, want nil", d2.Profile)
	}
	if !d2.IsPrivate {
		t.Error("d2.IsPrivate = false, want true")
	}

	// third device: wired
	d3 := devices[2]
	if d3.Wireless {
		t.Error("d3.Wireless = true, want false")
	}
}

func TestGetDeviceRaw(t *testing.T) {
	fixture := loadFixture(t, "devices.json")
	// Parse the fixture to get a single device for the raw response
	var resp APIResponse
	json.Unmarshal(fixture, &resp)
	var devices []json.RawMessage
	json.Unmarshal(resp.Data, &devices)

	singleDevice := []byte(`{"meta":{"code":200,"server_id":"srv-001","timestamp":1700000000},"data":` + string(devices[0]) + `}`)

	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/2.2/networks/12345/devices/aabbccdd1122" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Write(singleDevice)
	})

	raw, err := client.GetDeviceRaw("12345", "aabbccdd1122")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if raw == nil {
		t.Fatal("raw = nil")
	}
}

func TestUpdateDevice(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]interface{}

	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)
		w.Write(loadFixture(t, "empty_ok.json"))
	})

	err := client.UpdateDevice("12345", "dev1", map[string]interface{}{"paused": true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "PUT" {
		t.Errorf("Method = %q, want PUT", gotMethod)
	}
	if gotPath != "/2.2/networks/12345/devices/dev1" {
		t.Errorf("Path = %q, want /2.2/networks/12345/devices/dev1", gotPath)
	}
	if gotBody["paused"] != true {
		t.Errorf("body paused = %v, want true", gotBody["paused"])
	}
}

func TestPauseDevice(t *testing.T) {
	var gotBody map[string]interface{}
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)
		w.Write(loadFixture(t, "empty_ok.json"))
	})

	client.PauseDevice("net1", "dev1", true)
	if gotBody["paused"] != true {
		t.Errorf("paused = %v, want true", gotBody["paused"])
	}
}

func TestBlockDevice(t *testing.T) {
	var gotBody map[string]interface{}
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)
		w.Write(loadFixture(t, "empty_ok.json"))
	})

	client.BlockDevice("net1", "dev1", true)
	if gotBody["blocked"] != true {
		t.Errorf("blocked = %v, want true", gotBody["blocked"])
	}
}

func TestSetDeviceNickname(t *testing.T) {
	var gotBody map[string]interface{}
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)
		w.Write(loadFixture(t, "empty_ok.json"))
	})

	client.SetDeviceNickname("net1", "dev1", "My Device")
	if gotBody["nickname"] != "My Device" {
		t.Errorf("nickname = %v, want %q", gotBody["nickname"], "My Device")
	}
}

// --- Profiles ---

func TestGetProfiles(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/2.2/networks/12345/profiles" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Write(loadFixture(t, "profiles.json"))
	})

	profiles, err := client.GetProfiles("12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("len(profiles) = %d, want 2", len(profiles))
	}
	if profiles[0].Name != "Adults" {
		t.Errorf("profiles[0].Name = %q, want %q", profiles[0].Name, "Adults")
	}
	if profiles[1].Paused != true {
		t.Errorf("profiles[1].Paused = %v, want true", profiles[1].Paused)
	}
}

func TestGetProfileDetails(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.2/networks/12345/profiles/prof1" {
			t.Errorf("Path = %q", r.URL.Path)
		}
		w.Write(loadFixture(t, "profile_details.json"))
	})

	pd, err := client.GetProfileDetails("12345", "prof1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pd.Name != "Adults" {
		t.Errorf("Name = %q, want %q", pd.Name, "Adults")
	}
	if len(pd.Devices) != 1 {
		t.Fatalf("len(Devices) = %d, want 1", len(pd.Devices))
	}
	if pd.Devices[0].URL != "/2.2/networks/12345/devices/aabbccdd1122" {
		t.Errorf("Device URL = %q", pd.Devices[0].URL)
	}
}

func TestSetProfileDevices(t *testing.T) {
	var gotBody map[string]interface{}
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Method = %q, want PUT", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)
		w.Write(loadFixture(t, "empty_ok.json"))
	})

	client.SetProfileDevices("12345", "prof1", []string{"/2.2/networks/12345/devices/dev1"})
	devices, ok := gotBody["devices"].([]interface{})
	if !ok || len(devices) != 1 {
		t.Fatalf("devices = %v", gotBody["devices"])
	}
	d := devices[0].(map[string]interface{})
	if d["url"] != "/2.2/networks/12345/devices/dev1" {
		t.Errorf("device url = %v", d["url"])
	}
}

func TestPauseProfile(t *testing.T) {
	var gotBody map[string]interface{}
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)
		w.Write(loadFixture(t, "empty_ok.json"))
	})

	client.PauseProfile("net1", "prof1", true)
	if gotBody["paused"] != true {
		t.Errorf("paused = %v, want true", gotBody["paused"])
	}
}

// --- Eeros ---

func TestGetEeros(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/2.2/networks/12345/eeros" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Write(loadFixture(t, "eeros.json"))
	})

	eeros, err := client.GetEeros("12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(eeros) != 2 {
		t.Fatalf("len(eeros) = %d, want 2", len(eeros))
	}
	if eeros[0].Location != "Living Room" {
		t.Errorf("Location = %q, want %q", eeros[0].Location, "Living Room")
	}
	if !eeros[0].Gateway {
		t.Error("Gateway = false, want true")
	}
	if eeros[0].ConnectedClientsCount != 12 {
		t.Errorf("ConnectedClientsCount = %d, want 12", eeros[0].ConnectedClientsCount)
	}
	if eeros[1].MeshQualityBars != 3 {
		t.Errorf("MeshQualityBars = %d, want 3", eeros[1].MeshQualityBars)
	}
}

func TestRebootEero(t *testing.T) {
	var gotMethod, gotPath string
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Write(loadFixture(t, "empty_ok.json"))
	})

	err := client.RebootEero("8318690")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "POST" {
		t.Errorf("Method = %q, want POST", gotMethod)
	}
	if gotPath != "/2.2/eeros/8318690/reboot" {
		t.Errorf("Path = %q, want /2.2/eeros/8318690/reboot", gotPath)
	}
}

// --- Guest Network ---

func TestGetGuestNetwork(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.2/networks/12345/guestnetwork" {
			t.Errorf("Path = %q", r.URL.Path)
		}
		w.Write(loadFixture(t, "guest_network.json"))
	})

	gn, err := client.GetGuestNetwork("12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !gn.Enabled {
		t.Error("Enabled = false, want true")
	}
	if gn.Name != "Home Guest" {
		t.Errorf("Name = %q, want %q", gn.Name, "Home Guest")
	}
	if gn.Password != "guestpass123" {
		t.Errorf("Password = %q, want %q", gn.Password, "guestpass123")
	}
}

func TestEnableGuestNetwork(t *testing.T) {
	var gotBody map[string]interface{}
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)
		w.Write(loadFixture(t, "empty_ok.json"))
	})

	client.EnableGuestNetwork("12345", false)
	if gotBody["enabled"] != false {
		t.Errorf("enabled = %v, want false", gotBody["enabled"])
	}
}

func TestSetGuestNetworkPassword(t *testing.T) {
	var gotBody map[string]interface{}
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)
		w.Write(loadFixture(t, "empty_ok.json"))
	})

	client.SetGuestNetworkPassword("12345", "newpass")
	if gotBody["password"] != "newpass" {
		t.Errorf("password = %v, want %q", gotBody["password"], "newpass")
	}
}

// --- Reservations ---

func TestGetReservations(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.2/networks/12345/reservations" {
			t.Errorf("Path = %q", r.URL.Path)
		}
		w.Write(loadFixture(t, "reservations.json"))
	})

	reservations, err := client.GetReservations("12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reservations) != 2 {
		t.Fatalf("len(reservations) = %d, want 2", len(reservations))
	}
	if reservations[0].IP != "192.168.1.10" {
		t.Errorf("IP = %q, want %q", reservations[0].IP, "192.168.1.10")
	}
	if reservations[0].MAC != "11:22:33:44:55:66" {
		t.Errorf("MAC = %q, want %q", reservations[0].MAC, "11:22:33:44:55:66")
	}
	if reservations[0].Description != "NAS Server" {
		t.Errorf("Description = %q, want %q", reservations[0].Description, "NAS Server")
	}
}

func TestCreateReservation(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]interface{}
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)
		w.Write(loadFixture(t, "empty_ok.json"))
	})

	err := client.CreateReservation("12345", "192.168.1.50", "AA:BB:CC:DD:EE:FF", "Test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "POST" {
		t.Errorf("Method = %q, want POST", gotMethod)
	}
	if gotPath != "/2.2/networks/12345/reservations" {
		t.Errorf("Path = %q", gotPath)
	}
	if gotBody["ip"] != "192.168.1.50" {
		t.Errorf("ip = %v", gotBody["ip"])
	}
	if gotBody["mac"] != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("mac = %v", gotBody["mac"])
	}
	if gotBody["description"] != "Test" {
		t.Errorf("description = %v", gotBody["description"])
	}
}

func TestDeleteReservation(t *testing.T) {
	var gotMethod, gotPath string
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Write(loadFixture(t, "empty_ok.json"))
	})

	err := client.DeleteReservation("12345", "res1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "DELETE" {
		t.Errorf("Method = %q, want DELETE", gotMethod)
	}
	if gotPath != "/2.2/networks/12345/reservations/res1" {
		t.Errorf("Path = %q", gotPath)
	}
}

// --- Reboot network ---

func TestRebootNetwork(t *testing.T) {
	var gotMethod, gotPath string
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Write(loadFixture(t, "empty_ok.json"))
	})

	err := client.Reboot("12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "POST" {
		t.Errorf("Method = %q, want POST", gotMethod)
	}
	if gotPath != "/2.2/networks/12345/reboot" {
		t.Errorf("Path = %q", gotPath)
	}
}

// --- Login ---

func TestLogin(t *testing.T) {
	var gotBody map[string]interface{}
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/2.2/login" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)
		w.Write(loadFixture(t, "login.json"))
	})

	resp, err := client.Login("test@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.UserToken != "tok_abc123" {
		t.Errorf("UserToken = %q, want %q", resp.UserToken, "tok_abc123")
	}
	if gotBody["login"] != "test@example.com" {
		t.Errorf("login = %v, want %q", gotBody["login"], "test@example.com")
	}
}

func TestLoginVerify(t *testing.T) {
	var gotCookie string
	var gotBody map[string]interface{}
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotCookie = r.Header.Get("Cookie")
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)
		w.Write(loadFixture(t, "empty_ok.json"))
	})

	err := client.LoginVerify("user-token-123", "123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotCookie != "s=user-token-123" {
		t.Errorf("Cookie = %q, want %q", gotCookie, "s=user-token-123")
	}
	if gotBody["code"] != "123456" {
		t.Errorf("code = %v, want %q", gotBody["code"], "123456")
	}
}

// --- Error handling ---

func TestAPIError401(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(loadFixture(t, "error_401.json"))
	})

	_, err := client.GetAccount()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := err.Error(); got != "API error: unauthorized" {
		t.Errorf("error = %q, want %q", got, "API error: unauthorized")
	}
}

func TestAPIError500(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(loadFixture(t, "error_500.json"))
	})

	_, err := client.GetDevices("12345")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := err.Error(); got != "API error: internal server error" {
		t.Errorf("error = %q, want %q", got, "API error: internal server error")
	}
}

func TestAPIErrorMalformedJSON(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("not json"))
	})

	_, err := client.GetAccount()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestValidateTokenWithServer(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(loadFixture(t, "account.json"))
	})

	if !client.ValidateToken() {
		t.Error("ValidateToken() = false, want true")
	}
}

func TestValidateTokenFails(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(loadFixture(t, "error_401.json"))
	})

	if client.ValidateToken() {
		t.Error("ValidateToken() = true, want false")
	}
}
