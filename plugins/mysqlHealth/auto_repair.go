//go:build mysqlHealth

package main

import (
	"fmt"
	"os/exec"
	"time"

	lib "github.com/monobilisim/monokit2/lib"
	"github.com/rs/zerolog"
)

func AutoRepair(logger zerolog.Logger) {
	if !lib.DBConfig.Mysql.AutoRepair.Enabled {
		logger.Info().Msg("MySQL Auto Repair is disabled in configuration. Skipping auto repair.")
		return
	}

	if _, err := exec.LookPath("mysqlcheck"); err != nil {
		logger.Error().Err(err).Msg("mysqlcheck command not found. Please ensure mysql client tools are installed for auto repair functionality.")
		return
	}

	logger.Info().Msg("Starting MySQL auto repair process...")

	cronName := "mysql-auto-repair"
	// Mon Tue Wed Thu Fri Sat Sun
	dayString := lib.DBConfig.Mysql.AutoRepair.Day
	var day time.Weekday

	switch dayString {
	case "Mon":
		day = time.Monday
	case "Tue":
		day = time.Tuesday
	case "Wed":
		day = time.Wednesday
	case "Thu":
		day = time.Thursday
	case "Fri":
		day = time.Friday
	case "Sat":
		day = time.Saturday
	case "Sun":
		day = time.Sunday
	default:
		logger.Error().Msgf("Invalid day for MySQL auto repair: %s. Please use one of Mon, Tue, Wed, Thu, Fri, Sat, Sun.", dayString)
		return
	}

	if lib.IsTestMode() {
		day = time.Now().Weekday() // For testing, allow auto repair to run on any day
	}

	// 04:00 17:00 e.g.
	hourString := lib.DBConfig.Mysql.AutoRepair.Hour

	var hour time.Time
	hour, err := time.Parse("15:04", hourString)
	if err != nil {
		logger.Error().Err(err).Msgf("Invalid hour format for MySQL auto repair: %s. Please use HH:MM 24-hour format.", hourString)
		return
	}

	if lib.IsTestMode() {
		hour = time.Now() // For testing, allow auto repair to run at any time
	}

	var cronInterval lib.CronInterval
	cronInterval = lib.GetLastCronInterval(cronName)

	now := time.Now()

	if now.Weekday() != day {
		logger.Info().Msgf("Auto repair skipped: today is %s, scheduled day is %s.", now.Weekday(), day)
		return
	}

	if now.Hour() != hour.Hour() || now.Minute() < hour.Minute() {
		logger.Info().Msgf("Auto repair skipped: current time is %s, scheduled time is %s.", now.Format("15:04"), hour.Format("15:04"))
		return
	}

	if cronInterval.LastRun != nil {
		const minInterval = 6 * 24 * time.Hour
		if now.Sub(*cronInterval.LastRun) < minInterval {
			logger.Info().Msgf("Auto repair skipped: last run was %s, interval not yet reached.",
				cronInterval.LastRun.Format(time.RFC3339))
			return
		}
	}

	logger.Info().Msg("Starting MySQL auto repair process...")

	out, err := exec.Command(
		"mysqlcheck",
		"-h", lib.DBConfig.Mysql.Credentials.Host,
		"-P", fmt.Sprintf("%d", lib.DBConfig.Mysql.Credentials.Port),
		"-u", lib.DBConfig.Mysql.Credentials.User,
		fmt.Sprintf("-p%s", lib.DBConfig.Mysql.Credentials.Password),
		"--auto-repair",
		"--all-databases",
	).CombinedOutput()
	if err != nil {
		logger.Error().Err(err).Msgf("MySQL auto repair failed: %s", string(out))
		return
	}

	lib.CreateOrUpdateCronInterval(cronName)

	logger.Info().Msgf("MySQL auto repair completed successfully: %s", string(out))
}
