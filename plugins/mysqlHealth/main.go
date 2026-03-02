//go:build mysqlHealth

package main

import (
	"database/sql"
	"os"

	_ "github.com/go-sql-driver/mysql"
	mysql "github.com/go-sql-driver/mysql"
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

	mysqlConfig := mysql.NewConfig()
	mysqlConfig.User = lib.DBConfig.Mysql.Credentials.User
	mysqlConfig.Passwd = lib.DBConfig.Mysql.Credentials.Password
	mysqlConfig.Net = lib.DBConfig.Mysql.Credentials.Network
	mysqlConfig.Addr = lib.DBConfig.Mysql.Credentials.Host
	mysqlConfig.DBName = lib.DBConfig.Mysql.Credentials.DBName

	dbconn, err := sql.Open("mysql", mysqlConfig.FormatDSN())
	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to MySQL database")
		return
	}
	Connection = dbconn

	err = Connection.Ping()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to ping MySQL database")
		Connection.Close()
		Connection = nil
	}

	if Connection == nil {
		logger.Error().Msg("MySQL connection is not established. Exiting plugin.")
		return
	}

	CheckProcess(logger)
}
