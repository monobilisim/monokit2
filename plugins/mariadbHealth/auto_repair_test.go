//go:build mariadbHealth

package main

import (
	"testing"

	lib "github.com/monobilisim/monokit2/lib"
)

func TestAutoRepair(t *testing.T) {
	lib.InitConfig(configFiles...)
	lib.InitializeDatabase()

	logger, err := lib.InitLogger()
	if err != nil {
		t.Errorf("Failed to initialize logger: %v", err)
	}

	AutoRepair(logger)

	t.Log("AutoRepair executed successfully. Please verify MariaDB connection and configuration for expected behavior.")
}
