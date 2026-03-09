//go:build mariadbHealth

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
var pluginName string = "mariadbHealth"
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

	if !lib.DBConfig.MariaDB.Alarm.Enabled {
		logger.Info().Msg("MariaDB Health monitoring plugin is disabled in configuration. Exiting plugin.")
		return
	}

	logger.Info().Msg("Starting MariaDB Health monitoring plugin...")

	Connection, err = connectMariaDB(logger)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to establish MariaDB connection. Exiting plugin.")
		Connection = nil
		return
	}

	if Connection == nil {
		logger.Error().Msg("MariaDB connection is not established. Exiting plugin.")
		return
	}

	var isMariaDB bool
	var versionComment string
	err = Connection.QueryRow("SELECT @@version_comment").Scan(&versionComment)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to query MariaDB version. Exiting plugin.")
		return
	}

	if strings.Contains(strings.ToLower(versionComment), "mariadb") {
		isMariaDB = true
	}

	if !isMariaDB {
		logger.Error().Msg("Connected database does not appear to be MariaDB. Exiting plugin.")
		return
	}

	mariadbInDocker := IsMariaDBInDocker(logger)
	if mariadbInDocker {
		logger.Info().Msg("MariaDB appears to be running in Docker. This may affect connection methods and performance.")
	}

	CheckProcess(logger)

	if lib.DBConfig.MariaDB.AutoRepair.Enabled {
		AutoRepair(logger)
	}

	if lib.DBConfig.MariaDB.PMMAgent.Enabled {
		CheckPMM(logger)
	}

	if lib.DBConfig.MariaDB.Cluster.Enabled && lib.DBConfig.MariaDB.Cluster.ClusterType == "galera" {

		CheckClusterCertification(logger)

		CheckInaccessibleClusters(logger)

		CheckClusterSynced(logger)

		/*

			// Check cluster overall status
			CheckClusterStatus()

			// Check individual node status
			CheckNodeStatus()

			// Check Galera receive queue
			CheckReceiveQueue()

			// Check Galera flow control
			CheckFlowControl()

		*/
	}
}
