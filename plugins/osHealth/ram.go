//go:build osHealth
// +build osHealth

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

var pluginName string = "osHealth"
var moduleName string = "memory"
var up string = "up"
var down string = "down"

func CheckSystemRAM(logger zerolog.Logger) {

	logger.Info().Msg("Starting RAM monitoring...")

	vm, err := mem.VirtualMemory()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get virtual memory stats")
		return
	}

	if vm.UsedPercent >= float64(lib.OsHealthConfig.RamUsageAlarm.Limit) {
		stringifiedLimit := fmt.Sprintf("%.2f", lib.OsHealthConfig.RamUsageAlarm.Limit)
		stringifiedUsed := fmt.Sprintf("%.2f", vm.UsedPercent)
		stringifiedInterval := fmt.Sprintf("%d", lib.GlobalConfig.ZulipAlarm.Interval)

		alarmMessage := "[osHealth] - " + lib.GlobalConfig.Hostname + " - RAM usage has been more than " + stringifiedLimit + "% (" + stringifiedUsed + "%) for last " + stringifiedInterval + " minutes"

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
			alarmMessage := "[osHealth] - " + lib.GlobalConfig.Hostname + " - RAM usage is back to normal"

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
