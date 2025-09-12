package lib

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

var LogDir string
var DbDir string
var PluginsDir string
var GlobalConfig GlobalConfigType
var OsHealthConfig OsHealthConfigType

func InitConfig() error {
	if _, err := os.Stat("/etc/mono"); os.IsNotExist(err) {
		err := os.MkdirAll("/etc/mono", 0755)
		if err != nil {
			return fmt.Errorf("failed to create /etc/mono directory: %w", err)
		}
	}

	globalConfigExists := false
	if _, err := os.Stat("/etc/mono/global.yml"); err == nil {
		globalConfigExists = true
	} else {
		return fmt.Errorf("global configuration file does not exist")
	}

	if globalConfigExists {
		globalConfigData, err := os.ReadFile("/etc/mono/global.yml")
		if err != nil {
			return fmt.Errorf("failed to read global configuration file: %w", err)
		}

		err = yaml.Unmarshal(globalConfigData, &GlobalConfig)
		if err != nil {
			return fmt.Errorf("failed to parse global configuration file: %w", err)
		}
	}

	// /var/log/monokit2/monokit.log -> /var/log/monokit2
	LogDir = strings.Join(strings.Split(GlobalConfig.LogLocation, "/")[0:len(strings.Split(GlobalConfig.LogLocation, "/"))-1], "/")

	DbDir = strings.Join(strings.Split(GlobalConfig.SqliteLocation, "/")[0:len(strings.Split(GlobalConfig.SqliteLocation, "/"))-1], "/")

	PluginsDir = GlobalConfig.PluginsLocation

	osHealthConfigExists := false
	if _, err := os.Stat("/etc/mono/os.yml"); err == nil {
		osHealthConfigExists = true
	} else {
		return fmt.Errorf("os configuration file does not exist")
	}

	if osHealthConfigExists {
		osHealthConfigData, err := os.ReadFile("/etc/mono/os.yml")
		if err != nil {
			return fmt.Errorf("failed to read os configuration file: %w", err)
		}

		err = yaml.Unmarshal(osHealthConfigData, &OsHealthConfig)
		if err != nil {
			return fmt.Errorf("failed to parse os configuration file: %w", err)
		}
	}

	return nil
}
