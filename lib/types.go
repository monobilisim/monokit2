package lib

import "gorm.io/gorm"

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

type ZulipAlarm struct {
	gorm.Model
	Id                uint   `gorm:"primaryKey"`
	ProjectIdentifier string `gorm:"text"`
	Hostname          string `gorm:"text"`
	Content           string `gorm:"text"`
}

type Issue struct {
	gorm.Model
	Id           int    `gorm:"int" json:"id,omitempty"`
	Notes        string `gorm:"text" json:"notes,omitempty"`
	ProjectId    string `gorm:"text" json:"project_id,omitempty"`
	TrackerId    int    `gorm:"int" json:"tracker_id,omitempty"`
	Description  string `gorm:"text" json:"description,omitempty"`
	Subject      string `gorm:"text" json:"subject,omitempty"`
	PriorityId   int    `gorm:"int" json:"priority_id,omitempty"`
	StatusId     int    `gorm:"int" json:"status_id,omitempty"`
	AssignedToId string `gorm:"text" json:"assigned_to_id"`
	Service      string `gorm:"text" json:"-"` // not a JSON field, used internally
}

type RedmineIssue struct {
	Issue Issue `json:"issue"`
}
