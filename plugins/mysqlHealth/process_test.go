//go:build mysqlHealth

package main

import (
	"database/sql"
	"testing"
	"time"

	lib "github.com/monobilisim/monokit2/lib"
)

func TestCheckProcess(t *testing.T) {
	lib.InitConfig(configFiles...)
	lib.InitializeDatabase()

	logger, err := lib.InitLogger()
	if err != nil {
		t.Errorf("Failed to initialize logger: %v", err)
	}
	moduleName := "process"

	connectMySQL(logger)

	if Connection == nil {
		t.Errorf("MySQL connection is not established. Cannot run CheckProcess test.")
	}

	processThreshold := lib.DBConfig.Mysql.ProcessLimit

	inflateProcesses := make([]*sql.DB, processThreshold+1)
	for i := range inflateProcesses {
		db, err := connectMySQL(logger)
		if err != nil {
			t.Errorf("Failed to open extra connection %d: %v", i, err)
		}
		inflateProcesses[i] = db
	}

	t.Logf("Inflated MySQL processes to %d to test CheckProcess alarm triggering", len(inflateProcesses)+1)

	// connect again to ensure we have a valid connection for CheckProcess
	connectMySQL(logger)

	CheckProcess(logger)

	// Down alarm
	lastAlarm, err := lib.GetLastZulipAlarm(pluginName, moduleName)
	if err != nil {
		t.Errorf("Failed to retrieve last alarm: %v", err)
	}

	if lastAlarm.Status != down {
		t.Errorf("Expected alarm status to be '%s' when process count exceeds threshold, got '%s'", down, lastAlarm.Status)
	}

	lastIssue, err := lib.GetLastRedmineIssue(pluginName, moduleName)
	if err != nil {
		t.Errorf("Failed to retrieve last issue: %v", err)
	}

	if lastIssue.Status != down {
		t.Errorf("Expected issue status to be '%s' when process count exceeds threshold, got '%s'", down, lastIssue.Status)
	}

	for _, db := range inflateProcesses {
		if db != nil {
			db.Close()
		}
	}

	t.Logf("Closed inflated connections")

	time.Sleep(5 * time.Second)

	// UP alarm
	CheckProcess(logger)

	lastAlarm, err = lib.GetLastZulipAlarm(pluginName, moduleName)
	if err != nil {
		t.Errorf("Failed to retrieve last alarm: %v", err)
	}

	if lastAlarm.Status != up {
		t.Errorf("Expected alarm status to be '%s' when process count goes back below threshold, got '%s'", up, lastAlarm.Status)
	}

	lastIssue, err = lib.GetLastRedmineIssue(pluginName, moduleName)
	if err != nil {
		t.Errorf("Failed to retrieve last issue: %v", err)
	}

	if lastIssue.Status != up {
		t.Errorf("Expected issue status to be '%s' when process count goes back below threshold, got '%s'", up, lastIssue.Status)
	}
}
