//go:build mariadbHealth

package main

import (
	"database/sql"

	mysql "github.com/go-sql-driver/mysql"
	lib "github.com/monobilisim/monokit2/lib"
	"github.com/rs/zerolog"
)

func connectMariaDB(logger zerolog.Logger) (*sql.DB, error) {
	mariadbConfig := mysql.NewConfig()
	mariadbConfig.User = lib.DBConfig.MariaDB.Credentials.User
	mariadbConfig.Passwd = lib.DBConfig.MariaDB.Credentials.Password
	mariadbConfig.Net = lib.DBConfig.MariaDB.Credentials.Network
	mariadbConfig.Addr = lib.DBConfig.MariaDB.Credentials.Host
	mariadbConfig.DBName = lib.DBConfig.MariaDB.Credentials.DBName
	mariadbConfig.AllowNativePasswords = lib.DBConfig.MariaDB.Credentials.AllowNativePasswords

	dbconn, err := sql.Open("mysql", mariadbConfig.FormatDSN())
	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to MariaDB database")
		return nil, err
	}
	Connection = dbconn

	err = Connection.Ping()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to ping MariaDB database")
		Connection.Close()
		Connection = nil
	}

	if Connection == nil {
		logger.Error().Msg("MariaDB connection is not established with provided credentials. Trying unix socket connection...")
		mariadbConfig.Net = "unix"
		mariadbConfig.Addr = lib.DBConfig.MariaDB.Credentials.Socket

		dbconn, err = sql.Open("mysql", mariadbConfig.FormatDSN())
		if err != nil {
			logger.Error().Err(err).Msg("Failed to connect to MariaDB database via unix socket")
			return nil, err
		}
		Connection = dbconn

		err = Connection.Ping()
		if err != nil {
			logger.Error().Err(err).Msg("Failed to ping MariaDB database via unix socket")
			Connection.Close()
			Connection = nil
		}
	}

	return Connection, err
}
