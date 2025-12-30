You can duplicate "skeleton-plugin" and use it as a starting point.

Best real example for plugin development is osHealth plugin.

All database operations should be done via gorm models defined in lib types.go file and automigrated inside lib schema.go file.

If your plugin requires a new database table please add the model struct to types.go file and add automigration to schema.go file.

Do not import anything other than lib inside plugins.

Zulip Alarm and Redmine Issue functions has limiting and interval checking in their own. You don't need to reimplement them just provide the necessary fields in your plugin configuration struct.

Creating Zulip, Redmine alarms and News in a plugin:

```go
// Down means something bad happened like service is down
down := "down"
// Up means everything is back to normal
up := "up"

lib.InitConfig()

logger, err := lib.InitLogger()
if err != nil {
	fmt.Println("Logger init error:", err)
	return
}

logger.Error().Err(err).Msg("Logger initialized successfully")

lib.InitializeDatabase()

// Plugin name is usually defined on main.go file on each plugin
pluginName := "pluginName"
// Module name is for example zfsHealth, zfsCapacity, Disk etc.
// Module names should be unique per check to identify them correctly and avoid alarm spamming, usually a file should have a single global moduleName variable
moduleName := "moduleName"

alarmMessage := "Your alarm message here"

// When a service is down
// Send Zulip Alarm
lib.SendZulipAlarm(alarmMessage, pluginName, moduleName, down)
// Or you really need to know if it was sent successfully
err := lib.SendZulipAlarm(alarmMessage, pluginName, moduleName, down)
if err != nil {
	logger.Error().Err(err).Msg("Failed to send Zulip alarm")
} else {
	logger.Info().Msg("Zulip alarm sent successfully")
}
// Send Redmine Issue
lastIssue, err := lib.GetLastRedmineIssue(pluginName, moduleName)
if err != nil {
	lib.Logger.Error().Err(err).Msg("Failed to get last issue from database")
	return
}
// Because of how redmine api works we need to build issue from scratch to send updates or new issues
var issue lib.Issue
// Because how Redmine works make sure subject is always the same, for variable values use Description or Notes fields
// Description is used when creating a new issue
// Notes is used when updating an existing issue
issueSubject := fmt.Sprintf("%s için sistem yükü %.2f üstüne çıktı", lib.GlobalConfig.Hostname, loadLimit)

if lastIssue.Status == up {
// Last issue is up and now we have a down situation for same Subject, so we update the issue with Notes
	issue = lib.Issue{
		ProjectIdentifier: lib.GlobalConfig.ProjectIdentifier,
		Hostname:          lib.GlobalConfig.Hostname,
		Subject:           issueSubject,
		Notes:             fmt.Sprintf("Sorun devam ediyor, sistem yükü %.2f", loadAverage.Load1),
		StatusId:          lib.IssueStatus.Feedback,
		PriorityId:        lib.IssuePriority.Urgent,
		Service:           pluginName,
		Module:            moduleName,
		Status:            down,
	}
} else {
	issue = lib.Issue{
		ProjectIdentifier: lib.GlobalConfig.ProjectIdentifier,
		Hostname:          lib.GlobalConfig.Hostname,
		Subject:           issueSubject,
		Description:       alarmMessage,
		StatusId:          lib.IssueStatus.Feedback,
		PriorityId:        lib.IssuePriority.Urgent,
		Service:           pluginName,
		Module:            moduleName,
		Status:            down,
	}
}

// Subject matters if an issue created in last 6 hours occurs again it will be updated instead of creating a new one
err = lib.CreateRedmineIssue(issue)

// When you need to send the service is back up message

// First we get last alarm to check if its down or up to prevent sending duplicate up messages
lastAlarm, err := lib.GetLastAlarm(pluginName, moduleName)
if err != nil {
	logger.Error().Err(err).Msg("Failed to get last alarm")
}

// If the last alarm status was down, we send an up alarm
if lastAlarm.Status == down {
  lib.SendZulipAlarm("Service is back up", pluginName, moduleName, up)
}

// For Redmine issue we do the same check
lastIssue, err := lib.GetLastRedmineIssue(pluginName, moduleName)
if err != nil {
	logger.Error().Err(err).Msg("Failed to get last issue from database")
	return
}

if lastIssue.Status == down {
	issue := lib.Issue{
		ProjectIdentifier: lib.GlobalConfig.ProjectIdentifier,
		Hostname:          lib.GlobalConfig.Hostname,
		Subject:           issueSubject,
		Notes:             fmt.Sprintf("Sistem yükü normale döndü (%.2f)", loadAverage.Load1),
		PriorityId:        lib.IssuePriority.Urgent,
		StatusId:          lib.IssueStatus.Closed,
		Service:           pluginName,
		Module:            moduleName,
		Status:            up,
	}

	lib.CreateRedmineIssue(issue)
}
```

Always set plugin header correctly. Example:

```go
//go:build pluginName
```

Build plugins with this command after setting each files' header correctly:

```bash
cd ./plugins/osHealth && go build -tags osHealth -o ../bin/
```

While testing you can change monokit2 settings to make it work under user account without sudo permissions. Change the following settings in /etc/mono/global.yml:

global.yml
```yaml
log-location: "/home/user/Desktop/monokit2/logs/monokit2.log"
sqlite-location: "/home/user/Desktop/monokit2/logs/monokit2.db"
plugins-location: "/home/user/Desktop/monokit2/plugins/bin"
```

After writing the plugin if you want to create devel builds of it from CI/CD you need to change the following setting in .github/workflows/release.yml:

```yaml
      - name: Build binaries
        id: build
        run: |
          ACTIVE_PLUGINS=(osHealth pluginName)
```
