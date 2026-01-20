package main

import (
	"fmt"
	"os"
	"time"

	lib "github.com/monobilisim/monokit2/lib"
	"github.com/rs/zerolog"
)

// comes from -ldflags "-X 'main.version=version'" flag in ci build
var version string

func main() {
	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "--version", "-v", "version", "v", "--dependencies", "-d", "dependencies", "d":
			lib.HandleCommonPluginArgs(os.Args, version, []string{})
			return
		default:
			// continue execution
		}
	}

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

	if len(os.Args) == 1 {
		fmt.Println("Monokit2 - A modular system monitoring and management tool")
		fmt.Println("Use 'monokit2 --help' for more information.")
	}

	// reset command to delete database and logs before database init so no schema issues
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--help", "-h", "help", "h":
			fmt.Println("Usage: monokit2 [command] [options]")
			fmt.Println("reset               Delete monokit2's database and logs, --force is for without confirmation")
			fmt.Println("-i, --interactive   Launch the interactive TUI interface")
			fmt.Println("-v, --version       Display the current version of monokit2")
			fmt.Println("-d, --dependencies  List configuration file dependencies")
			fmt.Println("Usage: monokit2 [pluginName] [pluginOptions]")
			return
		}

		if os.Args[1] == "reset" {
			var SqliteExists bool
			var LogfileExists bool

			if _, err := os.Stat(lib.GlobalConfig.SqliteLocation); err == nil {
				SqliteExists = true
			}

			if _, err := os.Stat(lib.GlobalConfig.LogLocation); err == nil {
				LogfileExists = true
			}

			if !SqliteExists && !LogfileExists {
				return
			}

			if len(os.Args) == 2 {
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
					return
				} else {
					fmt.Println("Aborting...")
					return
				}
			}

			if os.Args[2] == "--force" {
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

	if lib.GlobalConfig.AutoUpdate.Enabled {
		logger.Info().Msg("Auto-update is enabled, checking for last update date...")
		intervalName := "auto-update"

		lastCronInterval := lib.GetLastCronInterval(intervalName)

		if lastCronInterval.LastRun == nil {
			update(intervalName, logger)
		}

		if !(lastCronInterval.LastRun == nil) {
			intervalInSeconds := lib.GlobalConfig.AutoUpdate.Interval * 60
			now := time.Now()

			if now.Sub(*lastCronInterval.LastRun).Seconds() < float64(intervalInSeconds) {
				logger.Info().Msg("Auto-update skipped due to interval not reached")
			} else {
				update(intervalName, logger)
			}
		}
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-i", "--interactive", "interactive", "i", "tui":
			if err := lib.RunTUI(version); err != nil {
				fmt.Printf("Error running TUI: %v\n", err)
				os.Exit(1)
			}
			return
		default:
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

}

func update(intervalName string, logger zerolog.Logger) {
	result, err := lib.UpdateMonokit2(version, false)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to update Monokit2")
	} else {
		logger.Info().Msg(fmt.Sprintf("%s updated from version %s to %s", result.Name, result.OldVersion, result.NewVersion))
	}

	results, err := lib.UpdatePlugins(version, false)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to update plugins")
	} else {
		for _, res := range results {
			logger.Info().Msg(fmt.Sprintf("Plugin %s updated from version %s to %s", res.Name, res.OldVersion, res.NewVersion))
		}
	}

	lib.CreateOrUpdateCronInterval(intervalName)
}
