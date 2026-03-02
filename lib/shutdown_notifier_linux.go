//go:build linux

package lib

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/rs/zerolog"
	"github.com/shirou/gopsutil/v4/host"
)

func HandleShutdownNotifier(args []string, logger zerolog.Logger) {
	if len(args) < 3 {
		logger.Error().Msg("shutdownNotifier requires 'up' or 'down' argument")
		return
	}

	action := args[2]
	if action != "up" && action != "down" {
		logger.Error().Str("action", action).Msg("Invalid shutdownNotifier argument. Use 'up' or 'down'.")
		return
	}

	uptime, _ := host.Uptime()
	duration := time.Duration(uptime) * time.Second
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60

	var uptimeStr string
	if hours > 24 {
		days := hours / 24
		hours = hours % 24
		uptimeStr = fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	} else {
		uptimeStr = fmt.Sprintf("%dh %dm", hours, minutes)
	}

	var message string
	if action == "down" {
		message = fmt.Sprintf("[%s] [:warning: Warning] Server is shutting down... Uptime: %s", GlobalConfig.Hostname, uptimeStr)
	} else if action == "up" {
		message = fmt.Sprintf("[%s] [:info: Info] Server is up...Uptime: %s", GlobalConfig.Hostname, uptimeStr)
	}

	err := SendZulipAlarm(message, "core", "power", action)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to send Zulip alarm for shutdownNotifier")
	} else {
		logger.Info().Str("action", action).Msg("Successfully sent shutdownNotifier alarm")
	}
}

func EnsureShutdownNotifierService(logger zerolog.Logger) {
	if IsTestMode() {
		return
	}

	_, err := exec.LookPath("systemctl")
	if err != nil {
		logger.Debug().Msg("systemctl not found, skipping shutdownNotifier service creation")
		return
	}

	servicePath := "/etc/systemd/system/shutdown-notifier.service"
	if _, err := os.Stat(servicePath); os.IsNotExist(err) {
		logger.Info().Msg("Creating shutdown-notifier.service as it does not exist")

		execPath, err := os.Executable()
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get executable path for shutdownNotifier")
			return
		}

		serviceContent := fmt.Sprintf(`[Unit]
Description=Notify shutdown&reboot process before reboot
After=network.target

[Service]
ExecStart=%s shutdownNotifier up
ExecStop=%s shutdownNotifier down
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
`, execPath, execPath)

		err = os.WriteFile(servicePath, []byte(serviceContent), 0644)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to write shutdown-notifier.service file (run as root?)")
			return
		}

		cmdReload := exec.Command("systemctl", "daemon-reload")
		if err := cmdReload.Run(); err != nil {
			logger.Error().Err(err).Msg("Failed to reload systemd daemon")
		}

		cmdEnable := exec.Command("systemctl", "enable", "--now", "shutdown-notifier.service")
		if err := cmdEnable.Run(); err != nil {
			logger.Error().Err(err).Msg("Failed to enable shutdown-notifier.service")
		} else {
			logger.Info().Msg("Successfully enabled and started shutdown-notifier.service")
		}
	}
}
