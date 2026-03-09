//go:build mariadbHealth

package main

import (
	"fmt"

	lib "github.com/monobilisim/monokit2/lib"
	"github.com/rs/zerolog"
)

func CheckInaccessibleClusters(logger zerolog.Logger) {
	moduleName := "cluster_inaccessible"

	rows, err := Connection.Query("SELECT @@wsrep_on")
	if err != nil {
		logger.Debug().Msgf("wsrep_on query failed: %v", err)
		return
	}
	defer rows.Close()

	var wsrepOn string
	if rows.Next() {
		err := rows.Scan(&wsrepOn)
		if err != nil {
			logger.Error().Err(err).Msg("Error scanning wsrep_on")
			return
		}
	}

	if wsrepOn != "ON" && wsrepOn != "1" {
		logger.Info().Msg("This node is not part of a Galera cluster. Skipping cluster accessibility checks.")
		return
	}

	rows, err = Connection.Query("SHOW GLOBAL STATUS WHERE Variable_name = 'wsrep_cluster_status'")
	if err != nil {
		logger.Error().Err(err).Msg("InaccessibleClusters query failed")
		return
	}
	defer rows.Close()

	var variableName, wsrepClusterStatus string
	if rows.Next() {
		if err := rows.Scan(&variableName, &wsrepClusterStatus); err != nil {
			logger.Error().Err(err).Msg("Error scanning rows")
			return
		}
	}

	rows, err = Connection.Query("SHOW STATUS WHERE Variable_name = 'wsrep_cluster_size'")
	if err != nil {
		logger.Error().Err(err).Msg("Error querying wsrep_cluster_size")
		return
	}
	defer rows.Close()

	var clusterSize string
	if rows.Next() {
		if err := rows.Scan(&variableName, &clusterSize); err != nil {
			logger.Error().Err(err).Msg("Error scanning wsrep_cluster_size")
			return
		}
	}

	logger.Debug().Msgf("wsrep_cluster_status: %s, wsrep_cluster_size: %s", wsrepClusterStatus, clusterSize)

	if wsrepClusterStatus != "Primary" {
		msg := fmt.Sprintf("[%s] - %s - Cluster status is not Primary, current Cluster status: %s", pluginName, lib.GlobalConfig.Hostname, wsrepClusterStatus)
		logger.Warn().Msg(msg)
		lib.SendZulipAlarm(msg, pluginName, moduleName, down)
	} else {
		lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleName)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to retrieve last cluster status alarm from database.")
		}

		if lastAlarm.Status == down {
			msg := fmt.Sprintf("[%s] - %s - Cluster status is Primary", pluginName, lib.GlobalConfig.Hostname)
			logger.Info().Msg(msg)
			lib.SendZulipAlarm(msg, pluginName, moduleName, up)
		}
	}

}
