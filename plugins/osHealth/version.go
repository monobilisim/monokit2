//go:build osHealth

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/monobilisim/monokit2/lib"
	vlib "github.com/monobilisim/monokit2/plugins/osHealth/vlib"
	"github.com/rs/zerolog"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
)

func CheckApplicationVersion(logger zerolog.Logger) {
	versionCheck := []string{"Docker", "Caddy", "Asterisk", "FrankenPHP", "HAProxy",
		"Jenkins", "MongoDB", "MySQL", "MariaDB", "Nginx",
		"OPNsense", "Postal", "PostgreSQL", "Redis", "Valkey",
		"Vault", "RabbitMQ", "Prometheus", "Zabbix", "PVE",
		"PMG", "PBS", "Zimbra"}

	logger.Info().Msg("Starting version monitoring...")

	// if version services are not installed for the applications, create empty records for them
	for _, app := range versionCheck {
		var appVersion []lib.Version
		err := lib.DB.Model(&lib.Version{}).Where("name = ?", app).Find(&appVersion).Error
		if err != nil {
			logger.Error().Err(err).Str("application", app).Msg("Error querying version from database")
			continue
		}
		if len(appVersion) == 0 {
			lib.DB.Create(&lib.Version{Name: app, Version: "", VersionMulti: "", Status: "not-installed"})
			continue
		}
	}

	vlib.DockerCheck(logger)
	vlib.CaddyCheck(logger)
	vlib.AsteriskCheck(logger)
	vlib.FrankenPHPCheck(logger)
	vlib.HAProxyCheck(logger)
	vlib.JenkinsCheck(logger)
	vlib.MariaDBCheck(logger)
	vlib.MongoDBCheck(logger)
	vlib.MySQLCheck(logger)
	vlib.NginxCheck(logger)
	vlib.OPNsenseCheck(logger)
	vlib.PostalCheck(logger)
	vlib.PostgreSQLCheck(logger)
	vlib.RedisCheck(logger)
	vlib.ValkeyCheck(logger)
	vlib.VaultCheck(logger)
	vlib.RabbitMQCheck(logger)
	vlib.PrometheusCheck(logger)
	vlib.ZabbixCheck(logger)
	vlib.ProxmoxVECheck(logger)
	vlib.ProxmoxMGCheck(logger)
	vlib.ProxmoxBSCheck(logger)
	vlib.ZimbraCheck(logger)
}

func GetAppVersionsForUI() ([][]string, bool) {
	var versions []lib.Version
	var rows [][]string
	hasError := false

	err := lib.DB.Model(&lib.Version{}).Find(&versions).Error
	if err != nil {
		return [][]string{{"Database Error.", "Unreadable", "ERROR"}}, true
	}

	for _, v := range versions {
		if v.Status == "not-installed" {
			continue
		}

		versionOutput := v.Version
		if versionOutput == "" {
			versionOutput = v.VersionMulti
		}
		if versionOutput == "" {
			versionOutput = "Unknown"
		}

		rows = append(rows, []string{v.Name, versionOutput, v.Status})
	}

	if len(rows) == 0 {
		return [][]string{{"System", "No Installed Applications.", "-"}}, hasError
	}

	return rows, hasError
}

func GetSystemRAMForUI() ([][]string, string) {
	var rows [][]string
	overallStatus := "s"

	vm, err := mem.VirtualMemory()
	if err != nil {
		return [][]string{{"System Error", "Failed to read RAM information"}},
			"e"
	}

	const GB = float64(1024 * 1024 * 1024)

	totalRAM := float64(vm.Total) / GB
	availableRAM := float64(vm.Available) / GB
	usedRAM := totalRAM - availableRAM

	// The original formula from your code reflecting the true usage rate:
	usagePercentage := float64(vm.Total-vm.Available) / float64(vm.Total) * 100

	// 3. Status check (Has the alarm limit been exceeded?)
	status := "Normal"
	if usagePercentage >= float64(lib.OsHealthConfig.RamUsageAlarm.Limit) {
		status = "⚠ CRITICAL (Limit Exceeded)"
		overallStatus = "w"
	}

	// 4. Fill our table rows with actual data (formatting with 2 decimal places)
	rows = append(rows, []string{"Total Memory", fmt.Sprintf("%.2f GB", totalRAM)})
	rows = append(rows, []string{"Used", fmt.Sprintf("%.2f GB", usedRAM)})
	rows = append(rows, []string{"Available", fmt.Sprintf("%.2f GB", availableRAM)})
	rows = append(rows, []string{"Usage Percentage", fmt.Sprintf("%% %.2f", usagePercentage)})
	rows = append(rows, []string{"Status", status})

	return rows, overallStatus
}

func GetSystemLoadForUI() ([][]string, string) {
	var rows [][]string
	overallStatus := "s"

	cpuCores, err := cpu.Counts(false)
	if err != nil {
		return [][]string{{"System Error", "Failed to read CPU core count"}}, "e"
	}
	loadAverage, err := load.Avg()
	if err != nil {
		return [][]string{{"System Error", "Failed to read system load information"}}, "e"
	}

	// Status Check (Has the alarm limit been exceeded?)
	// Using the exact alarm calculation formula from your original code:
	loadLimit := lib.OsHealthConfig.SystemLoadAlarm.LimitMultiplier * float64(cpuCores)

	status := "Stable"
	if loadAverage.Load1 >= loadLimit {
		status = "⚠ CRITICAL (Load Limit Exceeded)"
		overallStatus = "w"
	}

	//Fill our table rows with actual data
	rows = append(rows, []string{"Physical Cores", fmt.Sprintf("%d", cpuCores)})
	rows = append(rows, []string{"Current Load (Load 1m)", fmt.Sprintf("%.2f", loadAverage.Load1)})
	rows = append(rows, []string{"Average Load (Load 5m)", fmt.Sprintf("%.2f", loadAverage.Load5)})
	rows = append(rows, []string{"Critical Load Limit", fmt.Sprintf("%.2f", loadLimit)})
	rows = append(rows, []string{"Status", status})

	return rows, overallStatus
}

func GetSystemDiskForUI() ([][]string, string) {
	var rows [][]string
	overallStatus := "s"

	diskPartitions, err := disk.Partitions(true)
	if err != nil {
		return [][]string{{"System Error", "Failed to read disk partitions", "-", "-"}}, "e"
	}

	supportedFilesystems := []string{"ext4", "ext3", "ext2", "xfs", "btrfs", "fat32", "vfat"}

	// 3. Loop through each disk partition
	for _, partition := range diskPartitions {

		// Skip this disk if the filesystem is not supported (e.g., snap, tmpfs)
		if !slices.Contains(supportedFilesystems, partition.Fstype) {
			continue
		}

		// Get disk usage data
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue // Move to the next disk if it cannot be read
		}

		// Status check (Has the alarm limit been exceeded?)
		status := "Normal"
		if usage.UsedPercent > float64(lib.OsHealthConfig.DiskUsageAlarm.Limit) {
			status = "⚠ CRITICAL"
			overallStatus = "w"
		}

		// Convert values to GB/MB using your original 'formatBytes' function
		usageFormatted := fmt.Sprintf("%s / %s", formatBytes(usage.Used), formatBytes(usage.Total))
		percentage := fmt.Sprintf("%% %.1f", usage.UsedPercent)

		// Create our table row for this disk: [Mount Point, Usage, Percentage, Status]
		rows = append(rows, []string{partition.Mountpoint, usageFormatted, percentage, status})
	}

	// Eğer listelenecek hiçbir geçerli disk bulunamazsa
	if len(rows) == 0 {
		return [][]string{{"System", "No disks found in supported formats", "-", "-"}}, overallStatus
	}

	return rows, overallStatus
}

func GetSystemZFSForUI() ([][]string, string) {

	var rows [][]string
	overallStatus := "s"

	// 1. Search for the command first to prevent crashes on Windows or non-ZFS systems
	_, err := exec.LookPath("zpool")
	if err != nil {
		return [][]string{{"System", "ZFS (zpool) command not found", "-", "-"}}, "w"
	}

	// 2. Execute the ZFS command
	out, err := exec.Command("zpool", "list", "-H", "-o", "name,health,capacity").Output()
	if err != nil {
		return [][]string{{"System Error", "Failed to read ZFS data", "-", "-"}}, "w"
	}

	// 3. Read the output line by line and put it into the table
	lines := string(out)
	for _, line := range strings.Split(lines, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Split by spaces (e.g., "poolName ONLINE 0%")
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		poolName := fields[0]
		health := fields[1]
		capacityStr := fields[2]

		// Convert capacity to a number (e.g., "80%" -> 80)
		capacityClean := strings.TrimSuffix(capacityStr, "%")
		capacity, _ := strconv.Atoi(capacityClean)

		// Status check: If health is not "ONLINE" or usage limit is exceeded, write CRITICAL
		status := "Normal"
		if health != "ONLINE" || capacity >= lib.OsHealthConfig.DiskUsageAlarm.Limit {
			status = "⚠ CRITICAL"
			overallStatus = "w"
		}

		// Our table row: [Pool Name, Health, Capacity, Status]
		rows = append(rows, []string{poolName, health, capacityStr, status})
	}

	if len(rows) == 0 {
		return [][]string{{"System", "No registered ZFS Pool found", "-", "-"}}, overallStatus
	}

	return rows, overallStatus
}

func GetSystemPowerForUI() [][]string {
	var rows [][]string
	// 1. Get System Uptime
	uptime, err := host.Uptime()
	uptimeStr := "Unknown"

	if err == nil {
		duration := time.Duration(uptime) * time.Second
		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60

		if hours > 24 {
			days := hours / 24
			hours = hours % 24
			uptimeStr = fmt.Sprintf("%d days, %d hours, %d min", days, hours, minutes)
		} else {
			uptimeStr = fmt.Sprintf("%d hours, %d min", hours, minutes)
		}
	}

	rows = append(rows, []string{"System Uptime", uptimeStr})

	// 2. Scheduled Task / Shutdown
	action := "None"
	scheduledTime := "-"

	if runtime.GOOS == "linux" {
		// systemd shutdown check for Linux
		data, err := os.ReadFile("/run/systemd/shutdown/scheduled")
		if err == nil {
			lines := strings.Split(string(data), "\n")
			var usec int64
			for _, line := range lines {
				if strings.HasPrefix(line, "USEC=") {
					usecStr := strings.TrimPrefix(line, "USEC=")
					usec, _ = strconv.ParseInt(usecStr, 10, 64)
				}
				if strings.HasPrefix(line, "MODE=") {
					action = strings.TrimPrefix(line, "MODE=")
				}
			}
			if usec > 0 {
				scheduledTime = time.Unix(usec/1000000, 0).Format("2006-01-02 15:04:05")
			}
		}
	} else if runtime.GOOS == "windows" {
		// Safe check for Windows to avoid crashes
		cmd := exec.Command("schtasks", "/query", "/fo", "LIST", "/v")
		output, err := cmd.Output()
		if err == nil {
			outputStr := strings.ToLower(string(output))
			if strings.Contains(outputStr, "shutdown") {
				action = "shutdown"
			} else if strings.Contains(outputStr, "reboot") {
				action = "reboot"
			}
		}
	}

	if action == "" || action == "none" {
		action = "None"
	} else {
		action = strings.ToUpper(action)
	}

	rows = append(rows, []string{"Scheduled Action", action})

	if action != "NONE" {
		rows = append(rows, []string{"Action Time", scheduledTime})
	}

	return rows
}

func GetSystemdForUI() ([][]string, string) {
	var rows [][]string
	overallStatus := "s"

	// 1. OS check to prevent crashes due to dbus errors on Windows
	if runtime.GOOS != "linux" {
		return [][]string{{"System", "Systemd only works in a Linux environment", "-", "-"}}, "w"
	}

	// 2. Call the function from the original code to get the services
	services, err := GetServiceStatus()
	if err != nil {
		return [][]string{{"System Error", "Failed to read service statuses", "-", "-"}}, "w"
	}

	// 3. Apply the exact same filtering from the original code to find monitored services
	monitoredCount := 0
	for _, service := range services {
		matched := false
		for _, pattern := range lib.OsHealthConfig.ServiceHealthAlarm.Services {
			if match, _ := filepath.Match(pattern, strings.TrimSuffix(service.Name, ".service")); match {
				matched = true
				break
			}
		}

		if !matched {
			continue // Skip if it's an unmonitored service
		}

		monitoredCount++

		// 4. Uptime Formatting
		uptimeStr := "-"
		if service.ActiveState == "active" && service.Uptime > 0 {
			duration := time.Duration(service.Uptime) * time.Second
			hours := int(duration.Hours())
			minutes := int(duration.Minutes()) % 60
			if hours > 24 {
				days := hours / 24
				hours = hours % 24
				uptimeStr = fmt.Sprintf("%d days %d h %d m", days, hours, minutes)
			} else {
				uptimeStr = fmt.Sprintf("%d h %d m", hours, minutes)
			}
		}

		// 5. Determine Status (Health)
		status := "Normal"
		if service.ActiveState != "active" {
			status = fmt.Sprintf("⚠ CRITICAL (%s)", service.ActiveState)
			overallStatus = "w"
		}

		// Remove the ".service" extension for a cleaner look
		cleanName := strings.TrimSuffix(service.Name, ".service")
		rows = append(rows, []string{cleanName, service.ActiveState, uptimeStr, status})
	}

	if monitoredCount == 0 {
		return [][]string{{"System", "No active services found in the monitoring list", "-", "-"}}, overallStatus
	}

	return rows, overallStatus
}
