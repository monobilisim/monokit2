//go:build mysqlHealth

package main

import (
	"os/exec"
	"strings"

	"github.com/rs/zerolog"
)

func IsMysqlInDocker(logger zerolog.Logger) bool {
	if _, err := exec.LookPath("docker"); err == nil {
		out, err := exec.Command(
			"docker", "ps",
			"-a",
			"--format", "{{.Image}}",
		).Output()
		if err != nil {
			logger.Debug().Err(err).Msg("IsMysqlInDocker: docker ps failed, assuming not in Docker")
			return false
		}

		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			lower := strings.ToLower(line)
			if strings.Contains(lower, "mysql") {
				logger.Debug().Str("image", line).Msg("IsMysqlInDocker: detected via docker ps")
				return true
			}
		}
	}

	logger.Debug().Msg("IsMysqlInDocker: no Docker indicators found")
	return false
}
