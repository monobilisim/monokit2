//go:build valkeyHealth

package main

import (
	"testing"
)

func TestDetectValkey(t *testing.T) {
	if !DetectValkey() {
		t.Errorf("expected valkey to be detected")
	}
}
func TestIsValkeySentinel(t *testing.T) {
	setupValkeyTest(t)
	isSentinel := IsValkeySentinel()
	t.Logf("isSentinel: %v", isSentinel)
}
func TestCheckValkeyConnection(t *testing.T) {
	setupValkeyTest(t)
	err := CheckValkeyConnection()
	if err != nil {
		t.Fatalf("expected valkey connection to succeed, got error: %v", err)
	}
}
func TestCheckValkeySlaveCount(t *testing.T) {
	setupValkeyTest(t)
	isMaster, count, isSentinel, slaveOK := CheckValkeySlaveCount()
	t.Logf("isMaster: %v, count: %v, isSentinel: %v, slaveOK: %v", isMaster, count, isSentinel, slaveOK)
}
