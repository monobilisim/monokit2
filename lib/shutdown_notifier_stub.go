//go:build !linux

package lib

import (
	"github.com/rs/zerolog"
)

func HandleShutdownNotifier(args []string, logger zerolog.Logger) {
	// Not supported on non-Linux platforms
	logger.Debug().Msg("shutdownNotifier is not supported on this platform")
}

func EnsureShutdownNotifierService(logger zerolog.Logger) {
	// Not supported on non-Linux platforms
	logger.Debug().Msg("shutdownNotifier service generation is not supported on this platform")
}
