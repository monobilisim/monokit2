package main

import (
	"testing"

	lib "github.com/monobilisim/monokit2/lib"
)

func TestCheckProcess(t *testing.T) {
	lib.InitConfig(configFiles...)
	lib.InitializeDatabase()

	moduleName := "process"

	t.Log("CheckProcess test is not implemented yet.")
}
