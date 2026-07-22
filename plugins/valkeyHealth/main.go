//go:build valkeyHealth

package main

import (
	"fmt"
	"os"
	"strings"

	lib "github.com/monobilisim/monokit2/lib"
	"github.com/redis/go-redis/v9"
)

var version string
var pluginName string = "valkeyHealth"
var up string = "up"
var down string = "down"
var configFiles []string = []string{"db.yml"}
var Connection *redis.Client

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
	moduleName := "connection"
	lib.InitializeDatabase()
	logger.Info().Msg("Starting Valkey Health monitoring plugin...")

	if !DetectValkey() {
		logger.Warn().Msg("Valkey service not detected")
		if lib.DBConfig.Valkey.Alarm.Enabled && lib.GlobalConfig.ZulipAlarm.Enabled {
			lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleName)
			if err != nil || lastAlarm.Status != down {
				lib.SendZulipAlarm(
					"Valkey connection failed",
					pluginName,
					moduleName,
					down,
				)
			}
		}
		return
	}
	err = InitValkey()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to initialize Valkey")
		if lib.DBConfig.Valkey.Alarm.Enabled && lib.GlobalConfig.ZulipAlarm.Enabled {
			lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleName)
			if err != nil || lastAlarm.Status != down {
				alarmMessage := fmt.Sprintf("[%s] - %s - Valkey connection failed",
					pluginName,
					lib.GlobalConfig.Hostname,
				)
				lib.SendZulipAlarm(alarmMessage, pluginName, moduleName, down)
			}
		}
		return
	}
	if lib.DBConfig.Valkey.Alarm.Enabled && lib.GlobalConfig.ZulipAlarm.Enabled {
		lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleName)
		if err == nil && lastAlarm.Status == down {
			alarmMessage := fmt.Sprintf("[%s] - %s - Valkey connection restored",
				pluginName,
				lib.GlobalConfig.Hostname,
			)
			lib.SendZulipAlarm(alarmMessage, pluginName, moduleName, up)
		}
	}
	readable, writeable := TestValkeyReadWrite()
	moduleRead := "read"
	moduleWrite := "write"
	if !readable {
		logger.Error().Msg("Valkey read test failed")
		if lib.DBConfig.Valkey.Alarm.Enabled && lib.GlobalConfig.ZulipAlarm.Enabled {
			lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleRead)
			if err != nil || lastAlarm.Status != down {
				alarmMessage := fmt.Sprintf("[%s] - %s - Valkey read test failed",
					pluginName,
					lib.GlobalConfig.Hostname,
				)
				lib.SendZulipAlarm(alarmMessage, pluginName, moduleRead, down)
			}
		}
	} else {
		if lib.DBConfig.Valkey.Alarm.Enabled && lib.GlobalConfig.ZulipAlarm.Enabled {
			lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleRead)
			if err == nil && lastAlarm.Status == down {
				alarmMessage := fmt.Sprintf("[%s] - %s - Valkey read test restored",
					pluginName,
					lib.GlobalConfig.Hostname,
				)
				lib.SendZulipAlarm(alarmMessage, pluginName, moduleRead, up)
			}
		}
	}
	if !writeable {
		logger.Error().Msg("Valkey write test failed")
		if lib.DBConfig.Valkey.Alarm.Enabled && lib.GlobalConfig.ZulipAlarm.Enabled {
			lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleWrite)
			if err != nil || lastAlarm.Status != down {
				alarmMessage := fmt.Sprintf("[%s] - %s - Valkey write test failed",
					pluginName,
					lib.GlobalConfig.Hostname,
				)
				lib.SendZulipAlarm(alarmMessage, pluginName, moduleWrite, down)
			}
		}
	} else {
		if lib.DBConfig.Valkey.Alarm.Enabled && lib.GlobalConfig.ZulipAlarm.Enabled {
			lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleWrite)
			if err == nil && lastAlarm.Status == down {
				alarmMessage := fmt.Sprintf("[%s] - %s - Valkey write test restored",
					pluginName,
					lib.GlobalConfig.Hostname,
				)
				lib.SendZulipAlarm(alarmMessage, pluginName, moduleWrite, up)
			}
		}
	}
	isMaster := IsValkeyMaster()
	slaveCount := GetActualSlaveCount()
	clients := GetConnectedClients()
	uptime := FormatUptime(GetValkeyUptime())
	memory := GetUsedMemory()
	isSentinel := IsValkeySentinel()
	persistence := GetPersistenceMode()
	expectedSlaveCount := lib.DBConfig.Valkey.SlaveCount
	slaveOK := CheckSlaveCount(expectedSlaveCount)

	moduleSlave := "slave_count"
	if isSentinel && isMaster {
		if slaveOK {
			if lib.DBConfig.Valkey.Alarm.Enabled && lib.GlobalConfig.ZulipAlarm.Enabled {
				lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleSlave)
				if err == nil && lastAlarm.Status == down {
					alarmMessage := fmt.Sprintf(
						"[%s] - %s - Valkey slave count restored",
						pluginName,
						lib.GlobalConfig.Hostname,
					)
					lib.SendZulipAlarm(alarmMessage, pluginName, moduleSlave, up)
				}
			}
		} else {
			logger.Warn().
				Int("expected", expectedSlaveCount).
				Int("actual", slaveCount).
				Msg("Slave count mismatch")

			if lib.DBConfig.Valkey.Alarm.Enabled && lib.GlobalConfig.ZulipAlarm.Enabled {
				lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleSlave)
				if err != nil || lastAlarm.Status != down {
					alarmMessage := fmt.Sprintf(
						"[%s] - %s - Valkey slave count mismatch (Expected: %d, Actual: %d)",
						pluginName,
						lib.GlobalConfig.Hostname,
						expectedSlaveCount,
						slaveCount,
					)
					lib.SendZulipAlarm(alarmMessage, pluginName, moduleSlave, down)
				}
			}
		}
	}

	logger.Info().Msg("Valkey service detected.")
	var dashboard strings.Builder
	dashboard.WriteString(
		lib.Log(lib.InfoBadge, "Valkey health checks completed.\n\n"),
	)
	serviceStatus := "Running"
	if !isServiceActive("valkey.service") &&
		!isServiceActive("valkey-server.service") {
		serviceStatus = "Stopped"
	}
	role := "Slave"
	if isMaster {
		role = "Master"
	}
	if CheckRoleChange(role) {
		logger.Warn().
			Str("new_role", role).
			Msg("Valkey role changed")
	}
	sentinel := "Disabled"
	if isSentinel {
		sentinel = "Enabled"
	}
	readStatus := "FAILED"
	if readable {
		readStatus = "OK"
	}
	writeStatus := "FAILED"
	if writeable {
		writeStatus = "OK"
	}
	valkeyVersion := GetValkeyVersion()
	valkeyData := []lib.KV{
		{Key: "Service", Value: serviceStatus},
		{Key: "Version", Value: valkeyVersion},
		{Key: "Role", Value: role},
		{Key: "Sentinel", Value: sentinel},
		{Key: "Persistence", Value: persistence},
		{Key: "Readable", Value: readStatus},
		{Key: "Writeable", Value: writeStatus},
		{Key: "Connected Clients", Value: fmt.Sprintf("%d", clients)},
		{Key: "Connected Slaves", Value: fmt.Sprintf("%d", slaveCount)},
		{Key: "Slave Check", Value: map[bool]string{true: "OK", false: "FAILED"}[slaveOK]},
		{Key: "Uptime", Value: uptime},
		{Key: "Memory Usage", Value: memory},
	}
	dashboard.WriteString(
		lib.Log(lib.InfoBadge, "Valkey Status\n"),
	)
	dashboard.WriteString(
		lib.RenderKeyValueList(valkeyData),
	)
	dashboard.WriteString("\n")
	fmt.Println(
		lib.RenderPluginCard(
			"VALKEY HEALTH MONITOR",
			dashboard.String(),
		),
	)
}
