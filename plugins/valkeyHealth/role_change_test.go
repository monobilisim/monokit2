//go:build valkeyHealth

package main

import (
	"testing"
)

func TestCheckRoleChange(t *testing.T) {
	lastRole = ""
	if CheckRoleChange("Master") {
		t.Errorf("first role should not be considered as a change.")
	}
	if CheckRoleChange("Master") {
		t.Errorf("same role should not trigger a change.")
	}
	if !CheckRoleChange("Slave") {
		t.Errorf("expected role change from Master to Slave to be detected.")
	}
	if CheckRoleChange("Slave") {
		t.Errorf("same role should not trigger a change.")
	}
	if !CheckRoleChange("Master") {
		t.Errorf("expected role change from Slave to Master to be detected.")
	}
}
func TestIsValkeyMaster(t *testing.T) {
	setupValkeyTest(t)
	_ = IsValkeyMaster()
}
func TestGetActualSlaveCount(t *testing.T) {
	setupValkeyTest(t)
	count := GetActualSlaveCount()
	if count < 0 {
		t.Errorf("expected non-negative slave count, got %d", count)
	}
}
func TestCheckSlaveCount(t *testing.T) {
	setupValkeyTest(t)
	expected := GetActualSlaveCount()
	if !CheckSlaveCount(expected) {
		t.Errorf("expected slave count check to succeed, got %d", expected)
	}
}
