//go:build valkeyHealth

package main

import (
	"testing"

	lib "github.com/monobilisim/monokit2/lib"
)

// This test requires the valkey service to be stopped before running
func TestCheckValkeyConnectionDownAlarm(t *testing.T) {
	if !lib.IsTestMode() {
		t.Skip("requires TEST=1")
	}
	err := lib.InitConfig(configFiles...)
	if err != nil {
		t.Fatal("failed to init config: ", err)
	}
	lib.InitializeDatabase()
	lib.DBConfig.Valkey.Alarm.Enabled = true
	lib.GlobalConfig.ZulipAlarm.Enabled = true
	err = CheckValkeyConnection()
	if err == nil {
		t.Fatalf("expected valkey connection error, got nil")
	}
	alarm, err := lib.GetLastZulipAlarm(pluginName, "connection")
	if err != nil {
		t.Fatalf("failed to get last zulip alarm: %s", err)
	}
	if alarm.Status != down {
		t.Fatalf("expected alarm status to be down, got %s", alarm.Status)
	}
}
func TestCheckValkeyReadAlarm(t *testing.T) {
	if !lib.IsTestMode() {
		t.Skip("requires TEST=1")
	}
	setupValkeyTest(t)
	lib.DBConfig.Valkey.Alarm.Enabled = true
	lib.GlobalConfig.ZulipAlarm.Enabled = true
	err := lib.SendZulipAlarm("Valkey read test failed",
		pluginName,
		"read",
		down,
	)
	if err != nil {
		t.Fatalf("failed to send zulip alarm: %v", err)
	}
	readable, _ := CheckValkeyReadWrite()
	if !readable {
		t.Fatalf("expected valkey to be readable, got false")
	}
	alarm, err := lib.GetLastZulipAlarm(pluginName, "read")
	if err != nil {
		t.Fatalf("failed to get last zulip alarm: %v", err)
	}
	if alarm.Status != up {
		t.Fatalf("expected alarm status to be up, got %s", alarm.Status)
	}
}
func TestCheckValkeyWriteAlarm(t *testing.T) {
	if !lib.IsTestMode() {
		t.Skip("requires TEST=1")
	}
	setupValkeyTest(t)
	lib.DBConfig.Valkey.Alarm.Enabled = true
	lib.GlobalConfig.ZulipAlarm.Enabled = true
	err := lib.SendZulipAlarm("Valkey write test failed",
		pluginName,
		"write",
		down,
	)
	if err != nil {
		t.Fatalf("failed to send zulip alarm: %v", err)
	}
	_, writeable := CheckValkeyReadWrite()
	if !writeable {
		t.Fatalf("expected valkey to be writeable, got false")
	}
	alarm, err := lib.GetLastZulipAlarm(pluginName, "write")
	if err != nil {
		t.Fatalf("failed to get last zulip alarm: %v", err)
	}
	if alarm.Status != up {
		t.Fatalf("expected alarm status to be up, got %s", alarm.Status)
	}
}
