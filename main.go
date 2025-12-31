package main

import (
	"fmt"
	"os"

	lib "github.com/monobilisim/monokit2/lib"
)

// comes from -ldflags "-X 'main.version=version'" flag in ci build
var version string

func main() {
	lib.HandleCommonPluginArgs(os.Args, version, []string{})

	if lib.IsTestMode() {
		dir, err := os.UserHomeDir()
		if err != nil {
			panic("Failed to get user home directory: " + err.Error())
		}

		err = os.Mkdir(fmt.Sprintf("%s/test", dir), os.ModePerm)
		if err != nil {
			panic("Failed to create test directory: " + err.Error())
		}

		err = os.MkdirAll(fmt.Sprintf("%s/.local/bin", dir), os.ModePerm)
		if err != nil {
			panic("Failed to create local bin directory: " + err.Error())
		}
	}

	err := lib.InitConfig()
	if err != nil {
		fmt.Printf("Error initializing config: %v\n", err)
	}

	missingPlugins := lib.CheckPluginDependencies()
	if len(missingPlugins) > 0 {
		fmt.Println("The following plugins are missing dependencies and may not work correctly:")
		for _, plugin := range missingPlugins {
			fmt.Printf("- %s\n", plugin)
		}
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
		switch os.Args[1] {
		case "-i", "--interactive", "interactive", "i", "tui":
			if err := lib.RunTUI(version); err != nil {
				fmt.Printf("Error running TUI: %v\n", err)
				os.Exit(1)
			}
		default:
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

}
