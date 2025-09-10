//go:build osHealth
// +build osHealth

package main

import (
	"fmt"
	"sort"
	"time"

	lib "github.com/monobilisim/monokit2/lib"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/process"
)

var pluginName string = "osHealth"

func main() {
	lib.InitConfig()

	logger, err := lib.InitLogger()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	lib.InitializeDatabase()

	// err = os.MkdirAll(lib.LogDir+"/osHealth", os.ModePerm)
	// if err != nil {
	// 	logger.Fatal().Err(err).Msg("Failed to create osHealth log directory")
	// }

	logger.Info().Msg("Starting OS Health monitoring plugin...")

	if lib.OsHealthConfig.SystemLoadAlarm.Enabled {
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

					fmt.Printf("%-8s %-25s %-10s %-10s\n", "PID", "NAME", "CPU%", "RAM%")
					for i, u := range usages {
						if i >= lib.OsHealthConfig.SystemLoadAlarm.TopProcesses.Processes {
							break
						}
						fmt.Printf("%-8d %-25s %-10.2f %-10.2f\n", u.Pid, u.Name, u.CPU, u.RAM)
					}
				}
			}

			err := lib.SendZulipAlarm(alarmMessage, &pluginName)
			if err == nil {
				lib.DB.Create(&lib.ZulipAlarm{
					ProjectIdentifier: lib.GlobalConfig.ProjectIdentifier,
					Hostname:          lib.GlobalConfig.Hostname,
					Content:           alarmMessage,
					Service:           "osHealth",
					Status:            "down",
				})
			}
		} else {
			var lastAlarm lib.ZulipAlarm

			err := lib.DB.
				Where("project_identifier = ? AND hostname = ? AND service = ?",
					lib.GlobalConfig.ProjectIdentifier,
					lib.GlobalConfig.Hostname,
					"osHealth").
				Order("id DESC").
				Limit(1).
				Find(&lastAlarm).Error

			if err != nil {
				lib.Logger.Error().Err(err).Msg("Failed to get last alarm from database")
				return
			}

			if lastAlarm.Status == "down" {
				alarmMessage := "[osHealth] - " + lib.GlobalConfig.Hostname + " - System load is back to normal"

				err := lib.SendZulipAlarm(alarmMessage, &pluginName)
				if err == nil {
					lib.DB.Create(&lib.ZulipAlarm{
						ProjectIdentifier: lib.GlobalConfig.ProjectIdentifier,
						Hostname:          lib.GlobalConfig.Hostname,
						Content:           alarmMessage,
						Service:           "osHealth",
						Status:            "up",
					})
				}
			}
		}
	}
}

type ProcUsage struct {
	Pid  int32
	Name string
	CPU  float64
	RAM  float32
}
