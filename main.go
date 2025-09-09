package main

import (
	"fmt"
	"os"

	lib "github.com/monobilisim/monokit2/lib"
)

func main() {
	lib.InitConfig()

	fmt.Println(lib.LogDir)
	err := os.MkdirAll(lib.LogDir, os.ModePerm)
	if err != nil {
		panic("Failed to create log directory: " + err.Error())
	}

	if err = lib.InitializeDatabase(); err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		return
	}

	logger, err := lib.InitLogger()
	if err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
		return
	}

	logger.Info().Msg("Logger initialized successfully")

	logger.Info().Msg("Starting the Zulip alarm worker...")
	lib.StartZulipAlarmWorker()
}
