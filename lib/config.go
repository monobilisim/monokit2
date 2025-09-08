package lib

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var GlobalConfig GlobalConfigType

func InitConfig() error {
	// Parse global config
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

	return nil
}

func init() {
	if err := InitConfig(); err != nil {
		fmt.Printf("Warning: %v\n", err)
	}
}

type GlobalConfigType struct {
	ProjectIdentifier string `yaml:"project-identifier"`
	Hostname          string `yaml:"hostname"`
	LogLocation       string `yaml:"log-location"`

	ZulipAlarm struct {
		Enabled     bool     `yaml:"enabled"`
		Interval    int      `yaml:"interval"`
		WebhookUrls []string `yaml:"webhook-urls"`

		BotApi struct {
			Enabled    bool     `yaml:"enabled"`
			AlarmUrl   string   `yaml:"alarm-urls"`
			Email      string   `yaml:"email"`
			ApiKey     string   `yaml:"api-key"`
			UserEmails []string `yaml:"user-emails"`
		}
	} `yaml:"zulip-alarm"`

	Redmine struct {
		Enabled  bool   `yaml:"enabled"`
		ApiKey   string `yaml:"api-key"`
		Url      string `yaml:"url"`
		Interval int    `yaml:"interval"`
	} `yaml:"redmine"`
}
