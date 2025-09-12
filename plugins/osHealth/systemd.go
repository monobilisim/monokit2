//go:build osHealth

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/rs/zerolog"
)

type SystemdUnits struct {
	Name        string
	LoadState   string
	ActiveState string
	SubState    string
	Uptime      string
	Description string
}

func CheckSystemInit(logger zerolog.Logger) {

	logger.Info().Msg("systemctl command found, checking services...")

	services, err := GetServiceStatus()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get unit statuses")
		return
	}

	for _, service := range services {
		fmt.Println(service)
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
			status.Uptime = time.Since(startTime).Truncate(time.Second).String()
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}
