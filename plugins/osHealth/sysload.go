//go:build osHealth

package main

import (
	"fmt"
	"sort"
	"time"

	lib "github.com/monobilisim/monokit2/lib"
	"github.com/rs/zerolog"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/process"
)

func CheckSystemLoad(logger zerolog.Logger) {
	var moduleName string = "sysload"

	logger.Info().Msg("Starting System Load monitoring...")

	loadAverage, err := load.Avg()

	if err != nil {
		logger.Error().Err(err).Msg("Failed to get load average")
	}

	// Get the number of physical CPU cores NOT LOGICAL
	cpuCores, err := cpu.Counts(false)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get CPU core count")
	}

	if loadAverage.Load1 >= lib.OsHealthConfig.SystemLoadAlarm.LimitMultiplier*float64(cpuCores) {
		stringifiedLoadLimit := fmt.Sprintf("%.2f", lib.OsHealthConfig.SystemLoadAlarm.LimitMultiplier*float64(cpuCores))
		stringifiedLoad := fmt.Sprintf("%.2f", loadAverage.Load1)
		stringifiedInterval := fmt.Sprintf("%d", lib.GlobalConfig.ZulipAlarm.Interval)

		alarmMessage := "[osHealth] - " + lib.GlobalConfig.Hostname + " - System load has been more than " + stringifiedLoadLimit + " (" + stringifiedLoad + ")" + " for last " + stringifiedInterval + " minutes"

		if lib.OsHealthConfig.SystemLoadAlarm.TopProcesses.Enabled {
			processes, err := process.Processes()
			var usages []ProcUsage

			if err != nil {
				logger.Error().Err(err).Msg("Failed to get processes")
			}

			if err == nil {

				for _, p := range processes {
					_, _ = p.CPUPercent()
				}
				time.Sleep(time.Second)

				for _, p := range processes {
					cpu, err := p.CPUPercent()
					if err != nil {
						continue
					}

					mem, err := p.MemoryPercent()
					if err != nil {
						continue
					}

					name, _ := p.Name()

					usages = append(usages, ProcUsage{
						Pid:  p.Pid,
						Name: name,
						CPU:  cpu,
						RAM:  mem,
					})
				}

				sort.Slice(usages, func(i, j int) bool {
					return usages[i].CPU > usages[j].CPU
				})

				alarmMessage += "\n\nTop CPU consuming processes:\n"
				alarmMessage += fmt.Sprintf("%-8s %-25s %-10s %-10s\n", "PID", "NAME", "CPU%", "RAM%")
				for i, u := range usages {
					if i >= lib.OsHealthConfig.SystemLoadAlarm.TopProcesses.Processes {
						break
					}
					alarmMessage += fmt.Sprintf("%-8d %-25s %-10.2f %-10.2f\n", u.Pid, u.Name, u.CPU, u.RAM)
				}
			}
		}

		err := lib.SendZulipAlarm(alarmMessage, &pluginName, &moduleName, &down)
		if err == nil {
			lib.DB.Create(&lib.ZulipAlarm{
				ProjectIdentifier: lib.GlobalConfig.ProjectIdentifier,
				Hostname:          lib.GlobalConfig.Hostname,
				Content:           alarmMessage,
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
			lib.Logger.Error().Err(err).Msg("Failed to get last alarm from database")
			return
		}

		if lastAlarm.Status == down {
			alarmMessage := "[osHealth] - " + lib.GlobalConfig.Hostname + " - System load is back to normal"

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
