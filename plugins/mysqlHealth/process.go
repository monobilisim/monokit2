//go:build mysqlHealth

package main

import (
	"database/sql"
	"fmt"

	lib "github.com/monobilisim/monokit2/lib"
	"github.com/rs/zerolog"
)

func CheckProcess(logger zerolog.Logger) {
	var moduleName = "process"
	logger.Info().Msg("Checking MySQL processes...")

	rows, err := Connection.Query("SHOW PROCESSLIST")
	if err != nil {
		logger.Error().Err(err).Msg("Failed to execute SHOW PROCESSLIST query")
		return
	}
	defer rows.Close()

	var processCount int
	var processes []map[string]string
	for rows.Next() {
		var id int
		var user, host, db, command, time, state, info, timeMs sql.NullString

		err := rows.Scan(&id, &user, &host, &db, &command, &time, &state, &info, &timeMs)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to scan process row")
			continue
		}

		processCount++
		process := map[string]string{
			"id":      fmt.Sprintf("%d", id),
			"user":    user.String,
			"host":    host.String,
			"command": command.String,
			"time":    time.String,
			"state":   state.String,
		}
		processes = append(processes, process)
	}

	if err := rows.Err(); err != nil {
		logger.Error().Err(err).Msg("Error occurred during rows iteration")
		return
	}

	logger.Info().Msgf("Successfully retrieved MySQL processes. %d processes found.", processCount)
	logger.Debug().Interface("processes", processes).Msg("MySQL process details")

	processThreshold := lib.DBConfig.Mysql.ProcessLimit

	// Down alarm if process count is above threshold
	if lib.DBConfig.Mysql.Alarm.Enabled {
		if processCount > processThreshold {
			alarmMessage := fmt.Sprintf("[%s] - %s - MySQL process count has been more than the set limit %d, (%d)", pluginName, lib.GlobalConfig.Hostname, processThreshold, processCount)

			if lib.GlobalConfig.ZulipAlarm.Enabled {
				lib.SendZulipAlarm(alarmMessage, pluginName, moduleName, down)
			}

			/*
				if lib.GlobalConfig.Redmine.Enabled {
					lastIssue, err := lib.GetLastRedmineIssue(pluginName, moduleName)

					if err != nil {
						logger.Error().Err(err).Msg("Failed to get last Redmine issue")
					}

					var issue lib.Issue

					issueSubject := fmt.Sprintf("%s için MySQL process limiti aşıldı", lib.GlobalConfig.Hostname)

					if lastIssue.Status == up {
						issue = lib.Issue{
							ProjectIdentifier: lib.GlobalConfig.ProjectIdentifier,
							Hostname:          lib.GlobalConfig.Hostname,
							Notes:             "Sorun devam ediyor.",
							StatusId:          lib.IssueStatus.Feedback,
							PriorityId:        lib.IssuePriority.Urgent,
							Service:           pluginName,
							Module:            moduleName,
							Status:            down,
						}
					} else {
						issue = lib.Issue{
							ProjectIdentifier: lib.GlobalConfig.ProjectIdentifier,
							Hostname:          lib.GlobalConfig.Hostname,
							Subject:           issueSubject,
							Description:       alarmMessage,
							StatusId:          lib.IssueStatus.Feedback,
							PriorityId:        lib.IssuePriority.Urgent,
							Service:           pluginName,
							Module:            moduleName,
							Status:            down,
						}
					}

					lib.CreateRedmineIssue(issue)
				}
			*/
		}

		// UP alarm if process count is below threshold
		if processCount < processThreshold {
			alarmMessage := fmt.Sprintf("[%s] - %s - MySQL process count is back to normal (%d)", pluginName, lib.GlobalConfig.Hostname, processCount)

			if lib.GlobalConfig.ZulipAlarm.Enabled {
				lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleName)

				if err != nil {
					logger.Error().Err(err).Msg("Failed to get last Zulip alarm")
				}

				if lastAlarm.Status == down {
					lib.SendZulipAlarm(alarmMessage, pluginName, moduleName, up)
				}
			}

			/*
				if lib.GlobalConfig.Redmine.Enabled {
					lastIssue, err := lib.GetLastRedmineIssue(pluginName, moduleName)

					if err != nil {
						logger.Error().Err(err).Msg("Failed to get last Redmine issue")
					}

					if lastIssue.Status == down {
						issue := lib.Issue{
							ProjectIdentifier: lib.GlobalConfig.ProjectIdentifier,
							Hostname:          lib.GlobalConfig.Hostname,
							Subject:           fmt.Sprintf("%s için MySQL process limiti aşıldı", lib.GlobalConfig.Hostname),
							Description:       "Sorun çözüldü.",
							StatusId:          lib.IssueStatus.Resolved,
							PriorityId:        lib.IssuePriority.Urgent,
							Service:           pluginName,
							Module:            moduleName,
							Status:            up,
						}

						lib.CreateRedmineIssue(issue)
					}
				}
			*/
		}
	}
}
