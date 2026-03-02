//go:build mysqlHealth

package main

import (
	"fmt"

	"github.com/rs/zerolog"
)

func CheckProcess(logger zerolog.Logger) {
	rows, err := Connection.Query("SHOW PROCESSLIST")
	if err != nil {
		logger.Error().Err(err).Msg("Failed to execute SHOW PROCESSLIST query")
		return
	}
	defer rows.Close()

	fmt.Println("Current MySQL Processes:", rows)
}
