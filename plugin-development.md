You can duplicate "skeleton-plugin" and use it as a starting point.

Best real example for plugin development is osHealth plugin.

All database operations should be done via gorm models defined in lib types.go file and automigrated inside lib schema.go file.

If your plugin requires a new database table please add the model struct to types.go file and add automigration to schema.go file.

Do not import anything other than lib inside plugins.

Zulip Alarm and Redmine Issue functions has limiting and interval checking in their own. You don't need to reimplement them just provide the necessary fields in your plugin configuration struct.

Use zerolog for loggin inside plugins. [docs](https://github.com/rs/zerolog) Example:

```go

lib.InitConfig()

logger, err := lib.InitLogger()
if err != nil {
	fmt.Println("Logger init error:", err)
	return
}

logger.Error().Err(err).Msg("Logger initialized successfully")

lib.InitializeDatabase()

// after database is initialized you can use it like this
err := lib.SendZulipAlarm(alarmMessage, &pluginName, &moduleName, &down)

if err == nil {
	// available schemas are inside lib/schema.go and models are inside lib/types.go
	lib.DB.Create(&lib.ZulipAlarm{
		ProjectIdentifier: lib.GlobalConfig.ProjectIdentifier,
		Hostname:          lib.GlobalConfig.Hostname,
		Content:           alarmMessage,
		Service:           pluginName,
		Module:            moduleName,
		Status:            down,
	})
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
