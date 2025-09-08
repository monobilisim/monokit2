package main

import (
	"fmt"

	lib "github.com/monobilisim/monokit2/lib"
)

func main() {
	logger, err := lib.InitLogger()
	if err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
		return
	}

	logger.Info().Msg("Logger initialized successfully")
	fmt.Println(lib.GlobalConfig.ZulipAlarm.WebhookUrls)

	lib.SendZulipAlarm("test")
}
