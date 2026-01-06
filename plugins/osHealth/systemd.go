//go:build osHealth && linux

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/monobilisim/monokit2/lib"
	"github.com/rs/zerolog"
)

type SystemdUnits = lib.SystemdUnits

func CheckSystemInit(logger zerolog.Logger) {
	var moduleName string = "systemd"

	logger.Info().Msg("systemctl command found, checking services...")

	services, err := GetServiceStatus()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get unit statuses")
		return
	}

	for _, service := range services {

		var existingService SystemdUnits
		err := lib.DB.Model(&SystemdUnits{}).Where("name = ? AND project_identifier = ?", service.Name, lib.GlobalConfig.ProjectIdentifier).First(&existingService).Error
		if err != nil {
			logger.Error().Err(err).Msg("Failed to check if service exists in database")
		}

		if existingService.Name != service.Name {
			err := lib.DB.Create(&SystemdUnits{
				ProjectIdentifier: lib.GlobalConfig.ProjectIdentifier,
				Hostname:          lib.GlobalConfig.Hostname,
				Name:              service.Name,
				LoadState:         service.LoadState,
				ActiveState:       service.ActiveState,
				SubState:          service.SubState,
				Description:       service.Description,
				Uptime:            service.Uptime,
			}).Error

			if err != nil {
				logger.Error().Err(err).Msgf("Failed to insert %s into database", service.Name)
			}
			continue
		}

		var savedService SystemdUnits

		err = lib.DB.Model(&SystemdUnits{}).Where("name = ? AND project_identifier = ?", service.Name, lib.GlobalConfig.ProjectIdentifier).First(&savedService).Error
		if err != nil {
			logger.Error().Err(err).Msgf("Failed to get current service %s from database", service.Name)
			continue
		}

		// if service started in last 60 seconds that means it has restarted
		if savedService.Uptime > service.Uptime && service.Uptime > 0 && savedService.Uptime-service.Uptime < 60 {
			logger.Debug().Msgf("Service %s has restarted. Previous uptime: %d seconds, Current uptime: %d seconds", service.Name, savedService.Uptime, service.Uptime)
			alarmMessage := fmt.Sprintf("Service %s has restarted. Previous uptime: %d seconds, Current uptime: %d seconds", service.Name, savedService.Uptime, service.Uptime)

			lib.SendZulipAlarm(alarmMessage, pluginName, moduleName, down)

			if err == nil {
				err = lib.DB.Model(&lib.SystemdUnits{}).Where("name = ? AND project_identifier = ?", service.Name, lib.GlobalConfig.ProjectIdentifier).Updates(service).Error

				if err != nil {
					logger.Error().Err(err).Msgf("Failed to update service %s in database", service.Name)
				}
			}
		}

		if service.ActiveState != "active" && savedService.ActiveState == "active" {
			logger.Debug().Msgf("Service %s is down. Current state: %s", service.Name, service.ActiveState)
			alarmMessage := fmt.Sprintf("[osHealth] - %s - Service %s is down. Current state: %s", lib.GlobalConfig.Hostname, service.Name, service.ActiveState)

			err := lib.SendZulipAlarm(alarmMessage, pluginName, moduleName, down)
			if err == nil {
				err = lib.DB.Model(&lib.SystemdUnits{}).Where("name = ? AND project_identifier = ?", service.Name, lib.GlobalConfig.ProjectIdentifier).Updates(service).Error

				if err != nil {
					logger.Error().Err(err).Msgf("Failed to update service %s in database", service.Name)
				}
			}
		}

		if service.ActiveState == "active" && savedService.ActiveState != "active" {

			logger.Debug().Msgf("Service %s is active again. Current state: %s", service.Name, service.ActiveState)
			lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleName)

			if err != nil {
				logger.Error().Err(err).Msg("Failed to get last alarm from database")
				return
			}

			if lastAlarm.Status == down {
				alarmMessage := fmt.Sprintf("[osHealth] - %s - Service %s is now active", lib.GlobalConfig.Hostname, service.Name)

				err := lib.SendZulipAlarm(alarmMessage, pluginName, moduleName, up)
				if err == nil {
					err = lib.DB.Model(&lib.SystemdUnits{}).Where("name = ? AND project_identifier = ?", service.Name, lib.GlobalConfig.ProjectIdentifier).Updates(service).Error

					if err != nil {
						logger.Error().Err(err).Msgf("Failed to update service %s in database", service.Name)
					}
				}
			}
		}

	}
}

func GetServiceStatus() ([]SystemdUnits, error) {
	conn, err := dbus.New()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to dbus: %v", err)
	}
	defer conn.Close()

	units, err := conn.ListUnits()
	if err != nil {
		return nil, fmt.Errorf("failed to list units: %v", err)
	}

	var statuses []SystemdUnits

	for _, unit := range units {
		if !strings.HasSuffix(unit.Name, ".service") {
			continue
		}

		props, err := conn.GetUnitProperties(unit.Name)
		if err != nil {
			continue
		}

		status := SystemdUnits{
			Name:        unit.Name,
			LoadState:   unit.LoadState,
			ActiveState: unit.ActiveState,
			SubState:    unit.SubState,
			Description: unit.Description,
		}

		// Only calculate uptime if the service is active
		if ts, ok := props["ActiveEnterTimestamp"].(uint64); ok && unit.ActiveState == "active" && ts > 0 {
			startTime := time.Unix(0, int64(ts)*1000)
			status.Uptime = int64(time.Since(startTime).Seconds())
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}
