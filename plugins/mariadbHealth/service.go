//go:build mariadbHealth

package main

import (
	"os/exec"
	"strings"

	"github.com/rs/zerolog"
)

func IsMariaDBInDocker(logger zerolog.Logger) bool {
	if _, err := exec.LookPath("docker"); err == nil {
		out, err := exec.Command(
			"docker", "ps",
			"-a",
			"--format", "{{.Image}}",
		).Output()
		if err != nil {
			logger.Debug().Err(err).Msg("IsMariaDBInDocker: docker ps failed, assuming not in Docker")
			return false
		}

		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			lower := strings.ToLower(line)
			if strings.Contains(lower, "mariadb") {
				logger.Debug().Str("image", line).Msg("IsMariaDBInDocker: detected via docker ps")
				return true
			}
		}
	}

	logger.Debug().Msg("IsMariaDBInDocker: no Docker indicators found")
	return false
}
