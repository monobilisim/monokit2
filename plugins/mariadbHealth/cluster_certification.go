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

func GetClusterCertDataForUI() ([]lib.KV, string) {
	var count int
	var limiter int = 10
	overallStatus := "s"
	state := "OK"

	rows, err := Connection.Query("SELECT COUNT(*) FROM INFORMATION_SCHEMA.PROCESSLIST WHERE STATE LIKE '% for certificate%'")
	if err != nil {

		return []lib.KV{
			{Key: "CERTIFICATION METRIC", Value: "ERROR (Database query failed)"},
		}, "e"
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&count)
	}

	if count > limiter {
		state = "WARNING (Limit Exceeded)"
		overallStatus = "w"
	}

	return []lib.KV{
		{Key: "Waiting Certificates", Value: fmt.Sprintf("%d (Limit: %d)", count, limiter)},
		{Key: "Status", Value: state},
	}, overallStatus
}

func GetInaccessibleClusterDataForUI() ([]lib.KV, string) {
	overallStatus := "s"

	rows, err := Connection.Query("SELECT @@wsrep_on")
	if err != nil {
		return []lib.KV{{Key: "Galera Status", Value: "ERROR (wsrep_on query failed)"}}, "e"
	}
	defer rows.Close()

	var wsrepOn string
	if rows.Next() {
		rows.Scan(&wsrepOn)
	}

	if wsrepOn != "ON" && wsrepOn != "1" {
		return []lib.KV{
			{Key: "Galera Status", Value: "DISABLED (Not a Galera Node)"},
		}, "w"
	}

	rows, err = Connection.Query("SHOW GLOBAL STATUS WHERE Variable_name = 'wsrep_cluster_status'")
	if err != nil {
		return []lib.KV{{Key: "Cluster Status", Value: "ERROR (Status query failed)"}}, "e"
	}
	defer rows.Close()

	var variableName, wsrepClusterStatus string
	if rows.Next() {
		rows.Scan(&variableName, &wsrepClusterStatus)
	}

	rows, err = Connection.Query("SHOW STATUS WHERE Variable_name = 'wsrep_cluster_size'")
	if err != nil {
		return []lib.KV{
			{Key: "Cluster Status", Value: wsrepClusterStatus},
			{Key: "Cluster Size", Value: "ERROR (Size query failed)"},
		}, "e"
	}
	defer rows.Close()

	var clusterSize string
	if rows.Next() {
		rows.Scan(&variableName, &clusterSize)
	}

	statusDisplay := wsrepClusterStatus
	if wsrepClusterStatus != "Primary" {
		overallStatus = "e"
		statusDisplay = fmt.Sprintf("⚠ %s (Isolated Node)", wsrepClusterStatus)
	}

	return []lib.KV{
		{Key: "Galera Status", Value: "ON"},
		{Key: "Cluster Status", Value: statusDisplay},
		{Key: "Cluster Size", Value: clusterSize},
	}, overallStatus
}
