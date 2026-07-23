//go:build valkeyHealth

package main

import (
	"testing"

	lib "github.com/monobilisim/monokit2/lib"
)

func setupValkeyTest(t *testing.T) {
	t.Helper()
	err := lib.InitConfig()
	if err != nil {
		t.Fatalf("failed to init config: %v", err)
	}
	lib.InitializeDatabase()
	err = InitValkey()
	if err != nil {
		t.Fatalf("failed to init valkey: %v", err)
	}
}
func TestGetValkeyVersion(t *testing.T) {
	setupValkeyTest(t)
	version := GetValkeyVersion()
	if version == "" || version == "Unknown" {
		t.Fatal("failed to get valkey version")
	}
}
func TestGetConnectedClients(t *testing.T) {
	setupValkeyTest(t)
	clients := GetConnectedClients()
	if clients < 0 {
		t.Errorf("expected non-negative client count, got %d", clients)
	}
}
func TestGetValkeyUptime(t *testing.T) {
	setupValkeyTest(t)
	uptime := GetValkeyUptime()
	if uptime < 0 {
		t.Errorf("expected non-negative uptime, got %d", uptime)
	}
}
func TestGetUsedMemory(t *testing.T) {
	setupValkeyTest(t)
	usedMemory := GetUsedMemory()
	if usedMemory == "" || usedMemory == "Unknown" {
		t.Errorf("failed to get valkey used memory")
	}
}
func TestGetPersistenceMode(t *testing.T) {
	setupValkeyTest(t)
	mode := GetPersistenceMode()
	switch mode {
	case "RDB", "AOF", "RDB + AOF", "Disabled":
		// OK
	default:
		t.Errorf("unexpected persistence mode: %s", mode)
	}
}
func TestFormatUptime(t *testing.T) {
	tests := []struct {
		seconds  int
		expected string
	}{
		{0, "0d 0h 0m"},
		{60, "0d 0h 1m"},
		{3600, "0d 1h 0m"},
		{3661, "0d 1h 1m"},
		{90061, "1d 1h 1m"},
	}
	for _, test := range tests {
		got := FormatUptime(test.seconds)
		if got != test.expected {
			t.Errorf("FormatUptime(%d) = %q, expected %q",
				test.seconds, got, test.expected)
		}
	}
}
