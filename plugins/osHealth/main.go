//go:build osHealth
// +build osHealth

package main

import (
	"fmt"
	"os"

	lib "github.com/monobilisim/monokit2/lib"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/load"
)

func main() {
	lib.InitConfig()

	logger, err := lib.InitLogger()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	err = os.MkdirAll(lib.LogDir+"/osHealth", os.ModePerm)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create osHealth log directory")
	}

	logger.Info().Msg("Starting OS Health monitoring plugin...")

	if lib.OsHealthConfig.SystemLoadAlarm.Enabled {
		logger.Info().Msg("Starting System Load monitoring...")

		loadAverage, err := load.Avg()

		if err != nil {
			logger.Error().Err(err).Msg("Failed to get load average")
		}

		// Get the number of physical CPU cores NOT LOGICAL
		cpuCores, err1 := cpu.Counts(false)
		if err1 != nil {
			logger.Error().Err(err1).Msg("Failed to get CPU core count")
		}

		if loadAverage.Load1 >= lib.OsHealthConfig.SystemLoadAlarm.LimitMultiplier*float64(cpuCores) {
			stringifiedLoadLimit := fmt.Sprintf("%.2f", lib.OsHealthConfig.SystemLoadAlarm.LimitMultiplier*float64(cpuCores))
			stringifiedInterval := fmt.Sprintf("%d", lib.GlobalConfig.ZulipAlarm.Interval)

			alarmMessage := "[osHealth] - " + lib.GlobalConfig.Hostname + " - System load has been more than " + stringifiedLoadLimit + " for last " + stringifiedInterval + " minutes"

			lib.SendZulipAlarm(alarmMessage)

			lib.DB.Create(&lib.ZulipAlarm{
				ProjectIdentifier: lib.GlobalConfig.ProjectIdentifier,
				Hostname:          lib.GlobalConfig.Hostname,
				Content:           alarmMessage,
			})
		}
	}
}
