//go:build osHealth

package main

import (
	"github.com/monobilisim/monokit2/lib"
	vlib "github.com/monobilisim/monokit2/plugins/osHealth/vlib"
	"github.com/rs/zerolog"
)

func CheckApplicationVersion(logger zerolog.Logger) {
	versionCheck := []string{"Docker", "Caddy", "Asterisk", "FrankenPHP", "HAProxy", "Jenkins", "MongoDB", "MySQL"}
	logger.Info().Msg("Starting version monitoring...")

	// if version services are not installed for the applications, create empty records for them
	for _, app := range versionCheck {
		var appVersion []lib.Version
		err := lib.DB.Model(&lib.Version{}).Where("name = ?", app).Find(&appVersion).Error
		if err != nil {
			logger.Error().Err(err).Str("application", app).Msg("Error querying version from database")
			continue
		}
		if len(appVersion) == 0 {
			lib.DB.Create(&lib.Version{Name: app, Version: "", VersionMulti: "", Status: "not-installed"})
			continue
		}
	}

	vlib.DockerCheck(logger)
	vlib.CaddyCheck(logger)
	vlib.AsteriskCheck(logger)
	vlib.FrankenPHPCheck(logger)
	vlib.HAProxyCheck(logger)
	vlib.JenkinsCheck(logger)
	vlib.MySQLCheck(logger)
}
