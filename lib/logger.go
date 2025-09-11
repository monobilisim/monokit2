package lib

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

var (
	Logger zerolog.Logger
)

// default configured logger for the application
//
// logger, err := InitLogger() or lib.InitLogger()
func InitLogger() (zerolog.Logger, error) {
	var output *os.File
	var err error

	logLocation := GlobalConfig.LogLocation
	if logLocation == "" {
		logLocation = "stdout"
	}

	switch logLocation {
	case "stdout":
		output = os.Stdout
	default:
		output, err = os.OpenFile(logLocation, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return zerolog.Logger{}, fmt.Errorf("failed to open log file: %w", err)
		}
	}

	Logger = zerolog.New(output).With().Timestamp().Logger()
	return Logger, nil
}
