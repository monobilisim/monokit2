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
}

type ZulipAlarm struct {
	gorm.Model
	Id                uint   `gorm:"primaryKey"`
	ProjectIdentifier string `gorm:"text"` // internal use
	Hostname          string `gorm:"text"` // internal use
	Content           string `gorm:"text"`
	Status            string `gorm:"text"` // down or up
	Service           string `gorm:"text"` // plugin name
	Module            string `gorm:"text"` // plugin's module name
}

type Issue struct {
	gorm.Model
	Id                int     `gorm:"int" json:"id,omitempty"`
	Notes             string  `gorm:"text" json:"notes,omitempty"`
	ProjectId         string  `gorm:"text" json:"project_id,omitempty"`
	TrackerId         int     `gorm:"int" json:"tracker_id,omitempty"`
	Description       string  `gorm:"text" json:"description,omitempty"`
	Subject           string  `gorm:"text" json:"subject,omitempty"`
	PriorityId        int     `gorm:"int" json:"priority_id,omitempty"`
	StatusId          int     `gorm:"int" json:"status_id,omitempty"`
	AssignedToId      string  `gorm:"text" json:"assigned_to_id"`
	ProjectIdentifier string  `gorm:"text"`          // internal use
	Hostname          string  `gorm:"text"`          // internal use
	Status            *string `gorm:"text" json:"-"` // down or up
	Service           *string `gorm:"text" json:"-"` // plugin name
	Module            *string `gorm:"text" json:"-"` // plugin's module name
}

type RedmineIssue struct {
	Issue Issue `json:"issue"`
}

type SystemdUnits struct {
	gorm.Model
	id                uint   `gorm:"primaryKey"`
	ProjectIdentifier string `gorm:"text"` // internal use
	Hostname          string `gorm:"text"` // internal use
	Name              string `gorm:"text,unique"`
	LoadState         string `gorm:"text"`
	ActiveState       string `gorm:"text"`
	SubState          string `gorm:"text"`
	Uptime            int64  `gorm:"int"`
	Description       string `gorm:"text"`
}

var IssuePriority = struct {
	Default int
	Low     int
	Normal  int
	High    int
	Urgent  int
}{
	Default: 5, // 5 is the lowest priority
	Low:     4, // 4 is low priority
	Normal:  3, // 3 is normal priority
	High:    2, // 2 is high priority
	Urgent:  1, // 1 is the highest priority
}
