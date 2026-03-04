//go:build mysqlHealth

package main

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	lib "github.com/monobilisim/monokit2/lib"
)

func TestCheckPMM(t *testing.T) {
	lib.InitConfig(configFiles...)
	lib.InitializeDatabase()

	logger, err := lib.InitLogger()
	if err != nil {
		t.Errorf("Failed to initialize logger: %v", err)
	}

	moduleName := "pmm-agent"

	// Test PMM agent check with no actual agent installed (should skip without error)
	CheckPMM(logger)

	time.Sleep(3 * time.Second)

	lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleName)
	if err != nil {
		t.Errorf("Failed to retrieve last PMM alarm from database: %v", err)
	}

	fmt.Printf("Last PMM alarm: %v\n", lastAlarm)

	serviceContent := `[Unit]
Description=PMM-Agent Test Service

[Service]
ExecStart=/bin/sleep 3600
Type=simple

[Install]
WantedBy=multi-user.target
`

	t.Log("Creating dummy pmm-agent service")

	err = os.WriteFile("/etc/systemd/system/pmm-agent.service", []byte(serviceContent), 0644)
	if err != nil {
		t.Errorf("Failed to create service file: %v", err)
	}

	err = exec.Command("systemctl", "daemon-reload").Run()
	if err != nil {
		t.Errorf("Failed to reload systemd daemon: %v", err)
	}

	err = exec.Command("systemctl", "enable", "--now", "pmm-agent.service").Run()
	if err != nil {
		t.Errorf("Failed to start pmm-agent service: %v", err)
	}

	pmmAgentScript := `#!/bin/bash
# Dummy PMM agent script for testing
exit 0
`
	err = os.WriteFile("/usr/local/bin/pmm-agent", []byte(pmmAgentScript), 0755)
	if err != nil {
		t.Fatalf("Failed to create dummy pmm-agent script: %v", err)
	}
	defer os.Remove("/usr/local/bin/pmm-agent")

	t.Log("Created dummy pmm-agent binary")

	time.Sleep(5 * time.Second)

	CheckPMM(logger)

	lastAlarm, err = lib.GetLastZulipAlarm(pluginName, moduleName)
	if err != nil {
		t.Errorf("Failed to retrieve last PMM alarm from database: %v", err)
	}

	fmt.Printf("Last PMM alarm after starting service: %v\n", lastAlarm)

	// DOWN
	err = exec.Command("systemctl", "stop", "pmm-agent.service").Run()
	if err != nil {
		t.Errorf("Failed to stop pmm-agent service: %v", err)
	}

	time.Sleep(5 * time.Second)

	CheckPMM(logger)

	lastAlarm, err = lib.GetLastZulipAlarm(pluginName, moduleName)
	if err != nil {
		t.Errorf("Failed to retrieve last PMM alarm from database: %v", err)
	}

	if lastAlarm.Status != down {
		t.Errorf("Expected PMM alarm status '%s', got '%s'", down, lastAlarm.Status)
	}

	// UP
	err = exec.Command("systemctl", "start", "pmm-agent.service").Run()
	if err != nil {
		t.Errorf("Failed to start pmm-agent service: %v", err)
	}

	time.Sleep(5 * time.Second)

	CheckPMM(logger)

	lastAlarm, err = lib.GetLastZulipAlarm(pluginName, moduleName)
	if err != nil {
		t.Errorf("Failed to retrieve last PMM alarm from database: %v", err)
	}

	if lastAlarm.Status != up {
		t.Errorf("Expected PMM alarm status '%s', got '%s'", up, lastAlarm.Status)
	}
}
