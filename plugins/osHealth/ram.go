//go:build osHealth

package main

import (
	"fmt"
	"sort"
	"time"

	lib "github.com/monobilisim/monokit2/lib"
	"github.com/rs/zerolog"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
)

func CheckSystemRAM(logger zerolog.Logger) {
	var moduleName string = "memory"

	logger.Info().Msg("Starting RAM monitoring...")

	vm, err := mem.VirtualMemory()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get virtual memory stats")
		return
	}

	if vm.UsedPercent >= float64(lib.OsHealthConfig.RamUsageAlarm.Limit) {
		alarmMessage := fmt.Sprintf("[osHealth] - %s - RAM usage has been more than %d% (%.2f%) for last %d minutes", lib.GlobalConfig.Hostname, lib.OsHealthConfig.RamUsageAlarm.Limit, vm.UsedPercent, lib.GlobalConfig.ZulipAlarm.Interval)

		if lib.OsHealthConfig.RamUsageAlarm.TopProcesses.Enabled {
			processes, err := process.Processes()
			var usages []ProcUsage

			if err != nil {
				logger.Error().Err(err).Msg("Failed to get processes")
			} else {

				for _, p := range processes {
					_, _ = p.MemoryPercent()
				}
				time.Sleep(time.Second)

				for _, p := range processes {
					memPercent, err := p.MemoryPercent()
					if err != nil {
						continue
					}

					cpuPercent, _ := p.CPUPercent()
					name, _ := p.Name()

					usages = append(usages, ProcUsage{
						Pid:  p.Pid,
						Name: name,
						CPU:  cpuPercent,
						RAM:  memPercent,
					})
				}

				sort.Slice(usages, func(i, j int) bool {
					return usages[i].RAM > usages[j].RAM
				})

				alarmMessage += "\n\nTop RAM consuming processes:\n"
				alarmMessage += fmt.Sprintf("%-8s %-25s %-10s %-10s\n", "PID", "NAME", "CPU%", "RAM%")
				for i, u := range usages {
					if i >= lib.OsHealthConfig.RamUsageAlarm.TopProcesses.Processes {
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
			lib.Logger.Error().Err(err).Msg("Failed to get last RAM alarm from database")
			return
		}

		if lastAlarm.Status == down {
			alarmMessage := fmt.Sprintf("[osHealth] - %s - RAM usage is back to normal", lib.GlobalConfig.Hostname)

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
