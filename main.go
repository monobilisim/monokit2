package main

import (
	"fmt"
	"os"

	lib "github.com/monobilisim/monokit2/lib"
)

// comes from -ldflags "-X 'main.version=version'" flag in ci build
var version string

func main() {
	err := lib.InitConfig()
	if err != nil {
		fmt.Printf("Error initializing config: %v\n", err)
	}

	err = os.MkdirAll(lib.LogDir, os.ModePerm)
	if err != nil {
		panic("Failed to create log directory: " + err.Error())
	}

	// reset command to delete database and logs before database init so no schema issues
	if len(os.Args) > 1 {
		if os.Args[1] == "reset" || os.Args[1] == "reset --force" {

			if os.Args[1] == "reset" {
				fmt.Println("You are going to delete monokit2's database and logs. Are you sure? (y/n)")
				var response string
				fmt.Scanln(&response)
				if response == "y" || response == "Y" {
					if err := os.Remove(lib.GlobalConfig.SqliteLocation); err != nil {
						fmt.Printf("Error deleting database: %v\n", err)
						return
					}

					if err := os.RemoveAll(lib.GlobalConfig.LogLocation); err != nil {
						fmt.Printf("Error deleting logs: %v\n", err)
						return
					}
				} else {
					fmt.Println("Aborting...")
					return
				}
			}

			if os.Args[1] == "reset --force" {
				if err := os.Remove(lib.GlobalConfig.SqliteLocation); err != nil {
					fmt.Printf("Error deleting database: %v\n", err)
					return
				}

				if err := os.RemoveAll(lib.GlobalConfig.LogLocation); err != nil {
					fmt.Printf("Error deleting logs: %v\n", err)
					return
				}
			}

			return
		}
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

	err = os.MkdirAll(lib.PluginsDir, 0755)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create plugins directory")
	}

	if len(os.Args) > 1 {
		if os.Args[1] == "--version" || os.Args[1] == "-v" || os.Args[1] == "version" || os.Args[1] == "v" {
			fmt.Printf(version)
			return
		}

		// Run specific plugin with arguments
		plugin := os.Args[1]
		args := []string{}
		if len(os.Args) > 2 {
			args = os.Args[2:]
		}

		if err := lib.RunPlugin(plugin, args); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := lib.RunTUI(version); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
