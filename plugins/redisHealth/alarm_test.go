//go:build redisHealth

package main

import (
	"testing"

	lib "github.com/monobilisim/monokit2/lib"
)

// This test requires the Redis service to be stopped before running
func TestCheckRedisConnectionDownAlarm(t *testing.T) {
	if !lib.IsTestMode() {
		t.Skip("requires TEST=1")
	}
	err := lib.InitConfig(configFiles...)
	if err != nil {
		t.Fatal("failed to init config: ", err)
	}
	lib.InitializeDatabase()
	lib.DBConfig.Redis.Alarm.Enabled = true
	lib.GlobalConfig.ZulipAlarm.Enabled = true
	err = CheckRedisConnection()
	if err == nil {
		t.Fatalf("expected redis connection error, got nil")
	}
	alarm, err := lib.GetLastZulipAlarm(pluginName, "connection")
	if err != nil {
		t.Fatalf("failed to get last zulip alarm: %s", err)
	}
	if alarm.Status != down {
		t.Fatalf("expected alarm status to be down, got %s", alarm.Status)
	}
}
func TestCheckRedisReadAlarm(t *testing.T) {
	if !lib.IsTestMode() {
		t.Skip("requires TEST=1")
	}
	setupRedisTest(t)
	lib.DBConfig.Redis.Alarm.Enabled = true
	lib.GlobalConfig.ZulipAlarm.Enabled = true
	err := lib.SendZulipAlarm("Redis read test failed",
		pluginName,
		"read",
		down,
	)
	if err != nil {
		t.Fatalf("failed to send zulip alarm: %v", err)
	}
	readable, _ := CheckRedisReadWrite()
	if !readable {
		t.Fatalf("expected redis to be readable, got false")
	}
	alarm, err := lib.GetLastZulipAlarm(pluginName, "read")
	if err != nil {
		t.Fatalf("failed to get last zulip alarm: %v", err)
	}
	if alarm.Status != up {
		t.Fatalf("expected alarm status to be up, got %s", alarm.Status)
	}
}
func TestCheckRedisWriteAlarm(t *testing.T) {
	if !lib.IsTestMode() {
		t.Skip("requires TEST=1")
	}
	setupRedisTest(t)
	lib.DBConfig.Redis.Alarm.Enabled = true
	lib.GlobalConfig.ZulipAlarm.Enabled = true
	err := lib.SendZulipAlarm("Redis write test failed",
		pluginName,
		"write",
		down,
	)
	if err != nil {
		t.Fatalf("failed to send zulip alarm: %v", err)
	}
	_, writeable := CheckRedisReadWrite()
	if !writeable {
		t.Fatalf("expected redis to be writeable, got false")
	}
	alarm, err := lib.GetLastZulipAlarm(pluginName, "write")
	if err != nil {
		t.Fatalf("failed to get last zulip alarm: %v", err)
	}
	if alarm.Status != up {
		t.Fatalf("expected alarm status to be up, got %s", alarm.Status)
	}
}
