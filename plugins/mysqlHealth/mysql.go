//go:build mysqlHealth

package main

import (
	"database/sql"

	mysql "github.com/go-sql-driver/mysql"
	lib "github.com/monobilisim/monokit2/lib"
	"github.com/rs/zerolog"
)

func connectMySQL(logger zerolog.Logger) (*sql.DB, error) {
	mysqlConfig := mysql.NewConfig()
	mysqlConfig.User = lib.DBConfig.Mysql.Credentials.User
	mysqlConfig.Passwd = lib.DBConfig.Mysql.Credentials.Password
	mysqlConfig.Net = lib.DBConfig.Mysql.Credentials.Network
	mysqlConfig.Addr = lib.DBConfig.Mysql.Credentials.Host
	mysqlConfig.DBName = lib.DBConfig.Mysql.Credentials.DBName
	mysqlConfig.AllowNativePasswords = lib.DBConfig.Mysql.Credentials.AllowNativePasswords

	dbconn, err := sql.Open("mysql", mysqlConfig.FormatDSN())
	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to MySQL database")
		return nil, err
	}
	Connection = dbconn

	err = Connection.Ping()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to ping MySQL database")
		Connection.Close()
		Connection = nil
	}

	if Connection == nil {
		logger.Error().Msg("MySQL connection is not established with provided credentials. Trying unix socket connection...")
		mysqlConfig.Net = "unix"
		mysqlConfig.Addr = lib.DBConfig.Mysql.Credentials.Socket

		dbconn, err = sql.Open("mysql", mysqlConfig.FormatDSN())
		if err != nil {
			logger.Error().Err(err).Msg("Failed to connect to MySQL database via unix socket")
			return nil, err
		}
		Connection = dbconn

		err = Connection.Ping()
		if err != nil {
			logger.Error().Err(err).Msg("Failed to ping MySQL database via unix socket")
			Connection.Close()
			Connection = nil
		}
	}

	return Connection, err
}
