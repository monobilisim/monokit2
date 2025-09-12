//go:build osHealth

package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	lib "github.com/monobilisim/monokit2/lib"
	"github.com/rs/zerolog"
)

func CheckSystemDiskZFS(logger zerolog.Logger) {
	var moduleName string = "zfs"

	_, err := exec.LookPath("zpool")
	if err != nil {
		logger.Error().Err(err).Msg("zpool command not found")
		fmt.Println("zpool command not found")
		return
	}

	// monokit2-devel	ONLINE	0%
	out, err := exec.Command("zpool", "list", "-H", "-o", "name,health,capacity").Output()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to execute zpool command")
		return
	}

	// pool1 ONLINE 0%
	// pool2 DEGRADED 0%
	lines := string(out)
	for _, line := range strings.Split(lines, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// monokit2-devel ONLINE 0% => []string{"monokit2-devel", "ONLINE", "0%"}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		poolName := fields[0]
		health := fields[1]
		capacityStr := fields[2]

		if health != "ONLINE" {
			logger.Warn().Str("pool", poolName).Str("health", health).Msg("ZFS pool is not healthy")

			alarmMessage := fmt.Sprintf("[osHealth] - %s - ZFS pool %s is not healthy: %s", lib.GlobalConfig.Hostname, poolName, health)

			err := lib.SendZulipAlarm(alarmMessage, &pluginName, &moduleName, &down)
			if err == nil {
				lib.DB.Create(&lib.ZulipAlarm{
					ProjectIdentifier: lib.GlobalConfig.ProjectIdentifier,
					Hostname:          lib.GlobalConfig.Hostname,
					Content:           fmt.Sprintf("ZFS pool %s is not healthy: %s", poolName, health),
					Service:           pluginName,
					Module:            moduleName,
					Status:            down,
				})
			}
		} else {
			var lastAlarm lib.ZulipAlarm

			err := lib.DB.
				Where("project_identifier = ? AND hostname = ? AND service = ? AND module = ?",
					lib.GlobalConfig.ProjectIdentifier,
					lib.GlobalConfig.Hostname,
					pluginName,
					moduleName).
				Order("id DESC").
				Limit(1).
				Find(&lastAlarm).Error

			if err != nil {
				logger.Error().Err(err).Msg("Failed to get last alarm from database")
				return
			}

			if lastAlarm.Status == down {
				alarmMessage := fmt.Sprintf("[osHealth] - %s - ZFS pool %s is now healthy", lib.GlobalConfig.Hostname, poolName)

				err := lib.SendZulipAlarm(alarmMessage, &pluginName, &moduleName, &up)
				if err == nil {
					lib.DB.Create(&lib.ZulipAlarm{
						ProjectIdentifier: lib.GlobalConfig.ProjectIdentifier,
						Hostname:          lib.GlobalConfig.Hostname,
						Content:           alarmMessage,
						Service:           pluginName,
						Module:            moduleName,
						Status:            up,
					})
				}
			}
		}

		capacityStr = strings.TrimSuffix(capacityStr, "%")
		capacity, err := strconv.Atoi(capacityStr)
		if err != nil {
			logger.Error().Err(err).Str("capacity", capacityStr).Msg("Failed to parse capacity")
			continue
		}

		logger.Debug().
			Str("pool", poolName).
			Int("capacity_percent", capacity).
			Msg("ZFS pool usage information")

		moduleName := "zfsCapacity"
		if capacity >= lib.OsHealthConfig.DiskUsageAlarm.Limit {
			logger.Warn().Str("pool", poolName).Int("capacity", capacity).Msg("ZFS pool capacity exceeded limit")

			alarmMessage := fmt.Sprintf("[osHealth] - %s - ZFS pool %s capacity exceeded the limit %d%% (%s%%)", lib.GlobalConfig.Hostname, poolName, lib.OsHealthConfig.DiskUsageAlarm.Limit, capacityStr)

			err := lib.SendZulipAlarm(alarmMessage, &pluginName, &moduleName, &down)
			if err == nil {
				lib.DB.Create(&lib.ZulipAlarm{
					ProjectIdentifier: lib.GlobalConfig.ProjectIdentifier,
					Hostname:          lib.GlobalConfig.Hostname,
					Content:           fmt.Sprintf("ZFS pool %s is not healthy: %s", poolName, health),
					Service:           pluginName,
					Module:            moduleName,
					Status:            down,
				})
			}
		} else {
			var lastAlarm lib.ZulipAlarm

			err := lib.DB.
				Where("project_identifier = ? AND hostname = ? AND service = ? AND module = ?",
					lib.GlobalConfig.ProjectIdentifier,
					lib.GlobalConfig.Hostname,
					pluginName,
					moduleName).
				Order("id DESC").
				Limit(1).
				Find(&lastAlarm).Error

			if err != nil {
				logger.Error().Err(err).Msg("Failed to get last alarm from database")
				return
			}

			if lastAlarm.Status == down {
				alarmMessage := fmt.Sprintf("[osHealth] - %s - All ZFS pools are now under the limit of %d%%", lib.GlobalConfig.Hostname, lib.OsHealthConfig.DiskUsageAlarm.Limit)

				err := lib.SendZulipAlarm(alarmMessage, &pluginName, &moduleName, &up)
				if err == nil {
					lib.DB.Create(&lib.ZulipAlarm{
						ProjectIdentifier: lib.GlobalConfig.ProjectIdentifier,
						Hostname:          lib.GlobalConfig.Hostname,
						Content:           alarmMessage,
						Service:           pluginName,
						Module:            moduleName,
						Status:            up,
					})
				}
			}
		}
	}
}
