//go:build mariadbHealth

package main

import (
	"fmt"

	lib "github.com/monobilisim/monokit2/lib"
	"github.com/rs/zerolog"
)

func CheckClusterSynced(logger zerolog.Logger) {
	moduleName := "cluster_synced"

	rows, err := Connection.Query("SHOW GLOBAL STATUS WHERE Variable_name = 'wsrep_local_state_comment'")
	if err != nil {
		logger.Error().Err(err).Msg("CheckClusterSynced query failed")
		return
	}
	defer rows.Close()

	var variableName, wsrepLocalStateComment string
	if rows.Next() {
		if err := rows.Scan(&variableName, &wsrepLocalStateComment); err != nil {
			logger.Error().Err(err).Msg("Error scanning rows")
			return
		}
	}

	isSynced := wsrepLocalStateComment == "Synced"

	if !isSynced {
		msg := fmt.Sprintf("[%s] - %s - Cluster is not synced, state: %s", pluginName, lib.GlobalConfig.Hostname, wsrepLocalStateComment)
		lib.SendZulipAlarm(msg, pluginName, moduleName, down)
	} else {
		lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleName)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get last Zulip alarm for cluster synced")
		}

		if lastAlarm.Status == down {
			msg := fmt.Sprintf("[%s] - %s - Cluster is synced", pluginName, lib.GlobalConfig.Hostname)
			lib.SendZulipAlarm(msg, pluginName, moduleName, up)
		}
	}

}
func GetClusterSyncedDataForUI() ([]lib.KV, string) {
	overallStatus := "s"

	rows, err := Connection.Query("SHOW GLOBAL STATUS WHERE Variable_name = 'wsrep_local_state_comment'")
	if err != nil {
		return []lib.KV{{Key: "Sync Status", Value: "ERROR (Query failed)"}}, "e"
	}
	defer rows.Close()

	var variableName, wsrepLocalStateComment string
	if rows.Next() {
		rows.Scan(&variableName, &wsrepLocalStateComment)
	}

	displayState := wsrepLocalStateComment
	if wsrepLocalStateComment != "Synced" {
		overallStatus = "e"
		displayState = fmt.Sprintf("⚠ %s (Out of Sync)", wsrepLocalStateComment)
	}

	return []lib.KV{
		{Key: "Local State", Value: displayState},
	}, overallStatus
}
