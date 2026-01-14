package lib

type GlobalConfigType struct {
	ProjectIdentifier string `yaml:"project-identifier"`
	Hostname          string `yaml:"hostname"`
	LogLocation       string `yaml:"log-location"`
	SqliteLocation    string `yaml:"sqlite-location"`
	PluginsLocation   string `yaml:"plugins-location"`

	ZulipAlarm struct {
		Enabled     bool     `yaml:"enabled"`
		Interval    int      `yaml:"interval"`
		WebhookUrls []string `yaml:"webhook-urls"`
		Limit       int      `yaml:"limit"`

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
		Limit    int    `yaml:"limit"`
	} `yaml:"redmine"`

	AutoUpdate struct {
		Enabled  bool `yaml:"enabled"`
		Interval int  `yaml:"interval"`
	} `yaml:"auto-update"`
}

type OsHealthConfigType struct {
	SystemLoadAlarm struct {
		Enabled         bool    `yaml:"enabled"`
		LimitMultiplier float64 `yaml:"limit-multiplier"`

		TopProcesses struct {
			Enabled   bool `yaml:"enabled"`
			Processes int  `yaml:"processes"`
		} `yaml:"top-processes"`
	} `yaml:"system-load-alarm"`

	RamUsageAlarm struct {
		Enabled bool `yaml:"enabled"`
		Limit   int  `yaml:"limit"`

		TopProcesses struct {
			Enabled   bool `yaml:"enabled"`
			Processes int  `yaml:"processes"`
		} `yaml:"top-processes"`
	} `yaml:"ram-usage-alarm"`

	DiskUsageAlarm struct {
		Enabled bool `yaml:"enabled"`
		Limit   int  `yaml:"limit"`
	} `yaml:"disk-usage-alarm"`

	VersionAlarm struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"version-alarm"`
}

type UfwApplyConfigType struct {
	RuleSources []struct {
		Url      string `yaml:"url"`
		Protocol string `yaml:"protocol"`
		Port     string `yaml:"port"`
		Comment  string `yaml:"comment"`
	} `yaml:"rule-sources"`
	StaticRules []struct {
		IP       string `yaml:"ip"`
		Protocol string `yaml:"protocol"`
		Port     string `yaml:"port"`
		Comment  string `yaml:"comment"`
	} `yaml:"static-rules"`
	RulesetDir string `yaml:"ruleset-dir"`
}
