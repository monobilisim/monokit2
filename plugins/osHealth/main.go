//go:build osHealth

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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

	if lib.OsHealthConfig.DiskUsageAlarm.Enabled && hasZFS() {
		CheckSystemDiskZFS(logger)
	}
}

// checks if there is an active ZFS pool
func hasZFS() bool {
	_, err := exec.LookPath("zpool")
	if err != nil {
		return false
	}

	cmd := exec.Command("zpool", "list", "-H")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return len(strings.TrimSpace(string(output))) > 0
}
