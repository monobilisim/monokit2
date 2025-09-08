package lib

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func SendZulipAlarm(message string) error {
	if !GlobalConfig.ZulipAlarm.Enabled {
		Logger.Warn().Msg("Zulip alarm is bot enabled in the configuration")
		return nil
	}

	if GlobalConfig.ZulipAlarm.BotApi.Enabled {
		err := sendZulipBotApiAlarm(message)
		if err != nil {
			Logger.Error().Err(err).Msgf("failed to send Zulip Bot API alarm")
		}
	} else {
		for _, url := range GlobalConfig.ZulipAlarm.WebhookUrls {
			err := sendZulipWebhookAlarm(url, message)
			if err != nil {
				Logger.Error().Err(err).Msgf("Failed to send Zulip webhook alarm to %s", url)
			}
		}
	}

	return nil
}

// Placeholder for future implementation
func sendZulipBotApiAlarm(message string) error {

	return nil
}

func sendZulipWebhookAlarm(url string, message string) error {
	payload := map[string]string{
		"text": message,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		Logger.Error().Err(err).Msg("Failed to marshal Zulip webhook payload")
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		Logger.Error().Err(err).Msg("failed to send request")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		Logger.Error().Err(err).Msgf("Received non-2xx response: %d", resp.StatusCode)
	}

	return nil
}
