package lib

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

type OsHealthConfigType struct {
	Filesystems []string `yaml:"filesystems"`

	SystemLoadAlarm struct {
		Enabled         bool    `yaml:"enabled"`
		LimitMultiplier float64 `yaml:"limit-multiplier"`

		TopProcesses struct {
			Load struct {
				Enabled   bool `yaml:"enabled"`
				Processes int  `yaml:"processes"`
			} `yaml:"load"`
		} `yaml:"top-processes"`

		IssueInterval   int `yaml:"issue-interval"`
		IssueMultiplier int `yaml:"issue-multiplier"`
		IssueLimit      int `yaml:"issue-limit"`
	} `yaml:"system-load-alarm"`

	RamUsageAlarm struct {
		Enabled  bool `yaml:"enabled"`
		RamLimit int  `yaml:"ram-limit"`

		TopProcesses struct {
			Ram struct {
				Enabled   bool `yaml:"enabled"`
				Processes int  `yaml:"processes"`
			} `yaml:"ram"`
		} `yaml:"top-processes"`

		IssueInterval   int `yaml:"issue-interval"`
		IssueMultiplier int `yaml:"issue-multiplier"`
		IssueLimit      int `yaml:"issue-limit"`
	} `yaml:"ram-usage-alarm"`

	DiskUsageAlarm struct {
		Enabled          bool `yaml:"enabled"`
		DiskPartUseLimit int  `yaml:"disk-part-use-limit"`
	} `yaml:"disk-usage-alarm"`
}
