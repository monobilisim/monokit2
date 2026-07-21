//go:build osHealth && !linux

package main

import (
	"fmt"

	lib "github.com/monobilisim/monokit2/lib"
	"github.com/rs/zerolog"
)

func CheckSystemInit(logger zerolog.Logger) {
	return
}

type SystemdUnit = lib.SystemdUnit

func GetServiceStatus() ([]SystemdUnit, error) {
	return nil, fmt.Errorf("systemd not supported")
}
