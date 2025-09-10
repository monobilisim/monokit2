//go:build osHealth
// +build osHealth

package main

import (
	lib "github.com/monobilisim/monokit2/lib"
)

func main() {
	lib.InitConfig()

	logger, err := lib.InitLogger()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	lib.InitializeDatabase()

	logger.Info().Msg("Starting OS Health monitoring plugin...")

	if lib.OsHealthConfig.SystemLoadAlarm.Enabled {
		CheckSystemLoad(logger)
	}
}
