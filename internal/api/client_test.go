package api

import (
	"testing"
)

func TestExtractNetworkID(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"/2.2/networks/12345", "12345"},
		{"/2.2/networks/abc-def-ghi", "abc-def-ghi"},
	}

	for _, tt := range tests {
		result := ExtractNetworkID(tt.url)
		if result != tt.expected {
			t.Errorf("ExtractNetworkID(%q) = %q, want %q", tt.url, result, tt.expected)
		}
	}
}

func TestExtractDeviceID(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"/1664356/devices/f6af4e4424f1", "f6af4e4424f1"},
		{"/123/devices/abc-def", "abc-def"},
	}

	for _, tt := range tests {
		result := ExtractDeviceID(tt.url)
		if result != tt.expected {
			t.Errorf("ExtractDeviceID(%q) = %q, want %q", tt.url, result, tt.expected)
		}
	}
}

func TestExtractProfileID(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"/1664356/profiles/prof123", "prof123"},
		{"/123/profiles/abc-def", "abc-def"},
	}

	for _, tt := range tests {
		result := ExtractProfileID(tt.url)
		if result != tt.expected {
			t.Errorf("ExtractProfileID(%q) = %q, want %q", tt.url, result, tt.expected)
		}
	}
}

func TestDeviceDisplayName(t *testing.T) {
	tests := []struct {
		device   Device
		expected string
	}{
		{Device{Nickname: "My Phone", Hostname: "iphone", MAC: "aa:bb:cc:dd:ee:ff"}, "My Phone"},
		{Device{Hostname: "laptop", MAC: "11:22:33:44:55:66"}, "laptop"},
		{Device{MAC: "aa:bb:cc:dd:ee:ff"}, "aa:bb:cc:dd:ee:ff"},
	}

	for _, tt := range tests {
		result := tt.device.DisplayName()
		if result != tt.expected {
			t.Errorf("Device.DisplayName() = %q, want %q", result, tt.expected)
		}
	}
}

func TestNewClient(t *testing.T) {
	client := New("test-token")
	if client == nil {
		t.Fatal("New() returned nil")
	}

	if client.token != "test-token" {
		t.Errorf("client.token = %q, want %q", client.token, "test-token")
	}
}

func TestSetToken(t *testing.T) {
	client := New("initial-token")
	client.SetToken("new-token")

	if client.token != "new-token" {
		t.Errorf("client.token = %q, want %q", client.token, "new-token")
	}
}

func TestValidateTokenEmpty(t *testing.T) {
	client := New("")
	if client.ValidateToken() {
		t.Error("ValidateToken() should return false for empty token")
	}
}

func TestExtractEeroID(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"/2.2/eeros/12345", "12345"},
		{"/2.2/eeros/8318690", "8318690"},
		{"/2.2/eeros/abc-def-ghi", "abc-def-ghi"},
		{"8318690", "8318690"}, // Already just an ID
	}

	for _, tt := range tests {
		result := ExtractEeroID(tt.url)
		if result != tt.expected {
			t.Errorf("ExtractEeroID(%q) = %q, want %q", tt.url, result, tt.expected)
		}
	}
}

func TestShortenIPv6(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"fe80:0000:0000:0000:6a4a:76ff:fe06:318d", "fe80::6a4a:76ff:fe06:318d"},
		{"0000:0000:0000:0000:0000:0000:0000:0001", "::1"},
		{"2001:0db8:0000:0000:0000:0000:0000:0001", "2001:db8::1"},
		{"fe80:0:0:0:a9b:f1ff:fe25:9d4d", "fe80::a9b:f1ff:fe25:9d4d"}, // Already partially shortened
	}

	for _, tt := range tests {
		result := shortenIPv6(tt.input)
		if result != tt.expected {
			t.Errorf("shortenIPv6(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
