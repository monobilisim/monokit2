//go:build mysqlHealth

package main

import (
	"database/sql"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	lib "github.com/monobilisim/monokit2/lib"
)

// comes from -ldflags "-X 'main.version=version'" flag in ci build
var version string
var pluginName string = "mysqlHealth"
var up string = "up"
var down string = "down"
var configFiles []string = []string{"db.yml"}

var Connection *sql.DB

func main() {
	if len(os.Args) > 1 {
		lib.HandleCommonPluginArgs(os.Args, version, configFiles)
		return
	}

	err := lib.InitConfig(configFiles...)
	if err != nil {
		panic("Failed to initialize config: " + err.Error())
	}

	logger, err := lib.InitLogger()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	lib.InitializeDatabase()

	if !lib.DBConfig.Mysql.Alarm.Enabled {
		logger.Info().Msg("MySQL Health monitoring plugin is disabled in configuration. Exiting plugin.")
		return
	}

	logger.Info().Msg("Starting MySQL Health monitoring plugin...")

	Connection, err = connectMySQL(logger)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to establish MySQL connection. Exiting plugin.")
		Connection = nil
		return
	}

	if Connection == nil {
		logger.Error().Msg("MySQL connection is not established. Exiting plugin.")
		return
	}

	var isMysql bool
	var versionComment string
	err = Connection.QueryRow("SELECT @@version_comment").Scan(&versionComment)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to query MySQL version. Exiting plugin.")
		return
	}

	if strings.Contains(strings.ToLower(versionComment), "mysql") {
		isMysql = true
	}

	if !isMysql {
		logger.Error().Msg("Connected database does not appear to be MySQL. Exiting plugin.")
		return
	}

	mysqlInDocker := IsMysqlInDocker(logger)
	if mysqlInDocker {
		logger.Info().Msg("MySQL appears to be running in Docker. This may affect connection methods and performance.")
	}

	CheckProcess(logger)

	if lib.DBConfig.Mysql.AutoRepair.Enabled {
		AutoRepair(logger)
	}

	CheckPMM(logger)
}
