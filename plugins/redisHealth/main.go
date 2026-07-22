//go:build redisHealth

package main

import (
	"fmt"
	"os"
	"strings"

	lib "github.com/monobilisim/monokit2/lib"
	"github.com/redis/go-redis/v9"
)

var version string
var pluginName string = "redisHealth"
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
	lib.InitializeDatabase()
	logger.Info().Msg("Starting Redis Health monitoring plugin...")

	err = CheckRedisConnection()
	if err != nil {
		return
	}
	readable, writeable := CheckRedisReadWrite()

	isMaster, slaveCount, isSentinel, slaveOK := CheckRedisSlaveCount()
	clients := GetConnectedClients()
	uptime := FormatUptime(GetRedisUptime())
	memory := GetUsedMemory()
	persistence := GetPersistenceMode()
	moduleSlave := "slave_count"
	expectedSlaveCount := lib.DBConfig.Redis.SlaveCount

	if isSentinel && isMaster {
		if slaveOK {
			if lib.DBConfig.Redis.Alarm.Enabled && lib.GlobalConfig.ZulipAlarm.Enabled {
				lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleSlave)
				if err == nil && lastAlarm.Status == down {
					alarmMessage := fmt.Sprintf(
						"[%s] - %s - Redis slave count restored",
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

			if lib.DBConfig.Redis.Alarm.Enabled && lib.GlobalConfig.ZulipAlarm.Enabled {
				lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleSlave)
				if err != nil || lastAlarm.Status != down {
					alarmMessage := fmt.Sprintf(
						"[%s] - %s - Redis slave count mismatch (Expected: %d, Actual: %d)",
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

	logger.Info().Msg("Redis service detected.")
	var dashboard strings.Builder
	dashboard.WriteString(
		lib.Log(lib.InfoBadge, "Redis health checks completed.\n\n"),
	)
	serviceStatus := "Running"
	if !isServiceActive("redis.service") &&
		!isServiceActive("redis-server.service") {
		serviceStatus = "Stopped"
	}
	role := "Slave"
	if isMaster {
		role = "Master"
	}
	if CheckRoleChange(role) {
		logger.Warn().
			Str("new_role", role).
			Msg("Redis role changed")
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
	version := GetRedisVersion()
	redisData := []lib.KV{
		{Key: "Service", Value: serviceStatus},
		{Key: "Version", Value: version},
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
		lib.Log(lib.InfoBadge, "Redis Status\n"),
	)
	dashboard.WriteString(
		lib.RenderKeyValueList(redisData),
	)
	dashboard.WriteString("\n")
	fmt.Println(
		lib.RenderPluginCard(
			"REDIS HEALTH MONITOR",
			dashboard.String(),
		),
	)
}
func CheckRedisConnection() error {
	moduleName := "connection"
	if !DetectRedis() {
		lib.Logger.Warn().Msg("Redis service not detected")
		if lib.DBConfig.Redis.Alarm.Enabled && lib.GlobalConfig.ZulipAlarm.Enabled {
			lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleName)
			if err != nil || lastAlarm.Status != down {
				lib.SendZulipAlarm(
					"Redis connection failed",
					pluginName,
					moduleName,
					down,
				)
			}
		}
		return fmt.Errorf("redis service not detected")
	}
	err := InitRedis()
	if err != nil {
		lib.Logger.Error().Err(err).Msg("Failed to initialize Redis")

		if lib.DBConfig.Redis.Alarm.Enabled && lib.GlobalConfig.ZulipAlarm.Enabled {
			lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleName)
			if err != nil || lastAlarm.Status != down {
				alarmMessage := fmt.Sprintf(
					"[%s] - %s - Redis connection failed",
					pluginName,
					lib.GlobalConfig.Hostname,
				)
				lib.SendZulipAlarm(alarmMessage, pluginName, moduleName, down)
			}
		}
		return err
	}
	if lib.DBConfig.Redis.Alarm.Enabled && lib.GlobalConfig.ZulipAlarm.Enabled {
		lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleName)
		if err == nil && lastAlarm.Status == down {
			alarmMessage := fmt.Sprintf(
				"[%s] - %s - Redis connection restored",
				pluginName,
				lib.GlobalConfig.Hostname,
			)
			lib.SendZulipAlarm(alarmMessage, pluginName, moduleName, up)
		}
	}
	return nil
}
func CheckRedisReadWrite() (bool, bool) {
	readable, writeable := TestRedisReadWrite()

	moduleRead := "read"
	moduleWrite := "write"

	if !readable {
		lib.Logger.Error().Msg("Redis read test failed")
		if lib.DBConfig.Redis.Alarm.Enabled && lib.GlobalConfig.ZulipAlarm.Enabled {
			lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleRead)
			if err != nil || lastAlarm.Status != down {
				alarmMessage := fmt.Sprintf(
					"[%s] - %s - Redis read test failed",
					pluginName,
					lib.GlobalConfig.Hostname,
				)
				lib.SendZulipAlarm(alarmMessage, pluginName, moduleRead, down)
			}
		}
	} else {
		if lib.DBConfig.Redis.Alarm.Enabled && lib.GlobalConfig.ZulipAlarm.Enabled {
			lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleRead)
			if err == nil && lastAlarm.Status == down {
				alarmMessage := fmt.Sprintf(
					"[%s] - %s - Redis read test restored",
					pluginName,
					lib.GlobalConfig.Hostname,
				)
				lib.SendZulipAlarm(alarmMessage, pluginName, moduleRead, up)
			}
		}
	}
	if !writeable {
		lib.Logger.Error().Msg("Redis write test failed")
		if lib.DBConfig.Redis.Alarm.Enabled && lib.GlobalConfig.ZulipAlarm.Enabled {
			lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleWrite)
			if err != nil || lastAlarm.Status != down {
				alarmMessage := fmt.Sprintf(
					"[%s] - %s - Redis write test failed",
					pluginName,
					lib.GlobalConfig.Hostname,
				)
				lib.SendZulipAlarm(alarmMessage, pluginName, moduleWrite, down)
			}
		}
	} else {
		if lib.DBConfig.Redis.Alarm.Enabled && lib.GlobalConfig.ZulipAlarm.Enabled {
			lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleWrite)
			if err == nil && lastAlarm.Status == down {
				alarmMessage := fmt.Sprintf(
					"[%s] - %s - Redis write test restored",
					pluginName,
					lib.GlobalConfig.Hostname,
				)
				lib.SendZulipAlarm(alarmMessage, pluginName, moduleWrite, up)
			}
		}
	}
	return readable, writeable
}
func CheckRedisSlaveCount() (bool, int, bool, bool) {
	isMaster := IsRedisMaster()
	slaveCount := GetActualSlaveCount()
	isSentinel := IsRedisSentinel()
	expectedSlaveCount := lib.DBConfig.Redis.SlaveCount
	slaveOK := CheckSlaveCount(expectedSlaveCount)
	moduleSlave := "slave_count"

	if isSentinel && isMaster {
		if slaveOK {
			if lib.DBConfig.Redis.Alarm.Enabled && lib.GlobalConfig.ZulipAlarm.Enabled {
				lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleSlave)
				if err == nil && lastAlarm.Status == down {
					alarmMessage := fmt.Sprintf(
						"[%s] - %s - Redis slave count restored",
						pluginName,
						lib.GlobalConfig.Hostname,
					)
					lib.SendZulipAlarm(alarmMessage, pluginName, moduleSlave, up)
				}
			}
		} else {
			lib.Logger.Warn().
				Int("expected", expectedSlaveCount).
				Int("actual", slaveCount).
				Msg("Slave count mismatch")

			if lib.DBConfig.Redis.Alarm.Enabled && lib.GlobalConfig.ZulipAlarm.Enabled {
				lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleSlave)
				if err != nil || lastAlarm.Status != down {
					alarmMessage := fmt.Sprintf(
						"[%s] - %s - Redis slave count mismatch (Expected: %d, Actual: %d)",
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
	return isMaster, slaveCount, isSentinel, slaveOK
}
