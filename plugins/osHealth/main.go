//go:build osHealth

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	lib "github.com/monobilisim/monokit2/lib"
	"github.com/monobilisim/monokit2/ui"
)

// comes from -ldflags "-X 'main.version=version'" flag in ci build
var version string
var pluginName string = "osHealth"
var up string = "up"
var down string = "down"
var configFiles []string = []string{"os.yml"}

func main() {
	if len(os.Args) > 1 {
		lib.HandleCommonPluginArgs(os.Args, version, configFiles)
		return
	}

	err := lib.InitConfig(configFiles...)
	if err != nil {
		panic("Failed to initialize config: " + err.Error())
	}

	logger, err := lib.InitLogger()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	lib.InitializeDatabase()

	logger.Info().Msg("Starting OS Health monitoring plugin...")

	// checks supported application versions and reports when updated

	if lib.OsHealthConfig.VersionAlarm.Enabled {

		CheckApplicationVersion(logger)

	}

	// checks system load

	if lib.OsHealthConfig.SystemLoadAlarm.Enabled {

		CheckSystemLoad(logger)

	}

	// checks system RAM usage

	if lib.OsHealthConfig.RamUsageAlarm.Enabled {

		CheckSystemRAM(logger)

	}

	// checks system disk usage

	if lib.OsHealthConfig.DiskUsageAlarm.Enabled {

		CheckSystemDisk(logger)

	}

	// checks ZFS pool health and usage

	if lib.OsHealthConfig.DiskUsageAlarm.Enabled && hasZFS() {

		CheckSystemDiskZFS(logger)

	}

	// checks systemd services status

	if lib.OsHealthConfig.ServiceHealthAlarm.Enabled && hasSystemd() {

		CheckSystemInit(logger)

	}

	if lib.OsHealthConfig.PowerAlarm.Enabled {

		CheckSystemPowerHealth(logger)

	}

	var dashboard strings.Builder
	dashboard.WriteString(
		ui.Log(ui.InfoBadge, "Operating system health checks completed.\n\n"),
	)

	if lib.OsHealthConfig.VersionAlarm.Enabled {
		appRows, status := GetAppVersionsForUI()
		appHeaders := []string{"APPLICATION", "VERSION", "STATE"}

		badge := ui.SuccessBadge

		if status == false {
			badge = ui.ErrorBadge
		}
		dashboard.WriteString(ui.Log(badge, "Applications on the System:\n"))
		dashboard.WriteString(ui.RenderTable(appHeaders, appRows))
		dashboard.WriteString("\n")

	}

	if lib.OsHealthConfig.RamUsageAlarm.Enabled {
		ramRows, status := GetSystemRAMForUI()

		badge := ui.SuccessBadge

		if status == "w" {
			badge = ui.WarningBadge
		} else if status == "e" {
			badge = ui.ErrorBadge
		}

		dashboard.WriteString(ui.Log(badge, "System Memory Status:\n"))

		dashboard.WriteString(ui.RenderTable([]string{"PROPERTY", "VALUE"}, ramRows))

		dashboard.WriteString("\n\n")
	}

	if lib.OsHealthConfig.SystemLoadAlarm.Enabled {
		cpuRows, status := GetSystemLoadForUI()

		badge := ui.SuccessBadge

		if status == "w" {
			badge = ui.WarningBadge
		} else if status == "e" {
			badge = ui.ErrorBadge
		}

		dashboard.WriteString(
			ui.Log(badge, "Processor Load Status:\n"))

		dashboard.WriteString(
			ui.RenderTable(
				[]string{"PROPERTY", "VALUE"},
				cpuRows,
			))

		dashboard.WriteString("\n\n")
	}

	if lib.OsHealthConfig.DiskUsageAlarm.Enabled {
		diskRows, status := GetSystemDiskForUI()

		badge := ui.SuccessBadge

		if status == "w" {
			badge = ui.WarningBadge
		} else if status == "e" {
			badge = ui.ErrorBadge
		}

		dashboard.WriteString(
			ui.Log(badge, "System Disk Usage:\n"),
		)

		dashboard.WriteString(
			ui.RenderTable(
				[]string{"MOUNT POINT", "USED / TOTAL", "USAGE %", "STATUS"},
				diskRows,
			),
		)

		dashboard.WriteString("\n\n")
	}

	if lib.OsHealthConfig.DiskUsageAlarm.Enabled {
		zfsRows, status := GetSystemZFSForUI()
		badge := ui.SuccessBadge

		if status == "w" {
			badge = ui.WarningBadge
		} else if status == "e" {
			badge = ui.ErrorBadge
		}

		dashboard.WriteString(
			ui.Log(badge, "ZFS Pool Status:\n"),
		)

		dashboard.WriteString(
			ui.RenderTable(
				[]string{"POOL", "HEALTH", "CAPACITY", "STATUS"},
				zfsRows,
			),
		)

		dashboard.WriteString("\n\n")
	}

	if lib.OsHealthConfig.PowerAlarm.Enabled {

		powerRows := GetSystemPowerForUI()

		dashboard.WriteString(
			ui.Log(ui.WarningBadge, "System Power Status:\n"),
		)

		dashboard.WriteString(
			ui.RenderTable(
				[]string{"POWER / SYSTEM STATUS", "VALUE"},
				powerRows,
			),
		)

		dashboard.WriteString("\n\n")
	}

	if lib.OsHealthConfig.ServiceHealthAlarm.Enabled {
		if lib.OsHealthConfig.ServiceHealthAlarm.Enabled && hasSystemd() {

			systemdRows, status := GetSystemdForUI()
			badge := ui.SuccessBadge

			if status == "w" {
				badge = ui.WarningBadge
			} else if status == "e" {
				badge = ui.ErrorBadge
			}
			dashboard.WriteString(
				ui.Log(badge, "Monitored Services:\n"),
			)

			dashboard.WriteString(
				ui.RenderTable(
					[]string{"SERVICE NAME", "STATUS", "UPTIME", "HEALTH"},
					systemdRows,
				),
			)

			dashboard.WriteString("\n\n")
		}
	}

	fmt.Println(
		ui.RenderPluginCard(
			"OS HEALTH MONITOR",
			dashboard.String(),
		),
	)
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

// checks if systemd is available
func hasSystemd() bool {
	_, err := exec.LookPath("systemctl")
	if err != nil {
		return false
	}
	return true
}
