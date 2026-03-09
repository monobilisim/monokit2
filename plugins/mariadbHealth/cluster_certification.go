//go:build mariadbHealth

package main

import (
	"fmt"

	"github.com/monobilisim/monokit2/lib"
	"github.com/rs/zerolog"
)

func CheckClusterCertification(logger zerolog.Logger) {
	moduleName := "cluster_certification"
	var limiter int = 10

	rows, err := Connection.Query("SELECT COUNT(*) FROM INFORMATION_SCHEMA.PROCESSLIST WHERE STATE LIKE '% for certificate%'")
	if err != nil {
		logger.Error().Err(err).Msg("Failed to execute SELECT COUNT(*) FROM INFORMATION_SCHEMA.PROCESSLIST WHERE STATE LIKE '% for certificate%' query")
		return
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		rows.Scan(&count)
	}

	if count > limiter {
		msg := fmt.Sprintf("[%s] - %s - Certification waiting, limit: %d, count: %d", pluginName, lib.GlobalConfig.Hostname, limiter, count)
		lib.SendZulipAlarm(msg, pluginName, moduleName, down)
	}

	if count <= limiter {
		lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleName)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get last Zulip alarm for cluster certification")
		}

		if lastAlarm.Status == down {
			msg := fmt.Sprintf("[%s] - %s - Certification waiting OK, limit: %d, count: %d", pluginName, lib.GlobalConfig.Hostname, limiter, count)
			lib.SendZulipAlarm(msg, pluginName, moduleName, up)
		}
	}
}
