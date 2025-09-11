//go:build osHealth

package main

import (
	"fmt"
	"os"

	lib "github.com/monobilisim/monokit2/lib"
)

// comes from -ldflags "-X 'main.version=version'" flag in ci build
var version string
var pluginName string = "osHealth"
var up string = "up"
var down string = "down"

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "--version" || os.Args[1] == "-v" || os.Args[1] == "version" || os.Args[1] == "v" {
			fmt.Printf(version)
			return
		}
	}

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

	if lib.OsHealthConfig.RamUsageAlarm.Enabled {
		CheckSystemRAM(logger)
	}

	if lib.OsHealthConfig.DiskUsageAlarm.Enabled {
		CheckSystemDisk(logger)
	}
}
