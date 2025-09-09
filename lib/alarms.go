package lib

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

var (
	zulipAlarmQueue []string
	queueMutex      sync.Mutex
)

// This function actually queues the messages to be sent for not overwhelm the Zulip server when there are many alerts at the same time.
func SendZulipAlarm(message string) {
	queueMutex.Lock()
	zulipAlarmQueue = append(zulipAlarmQueue, message)
	queueMutex.Unlock()
}

func StartZulipAlarmWorker() {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			processQueue()
		}
	}()
}

func processQueue() {
	interval := time.Duration(GlobalConfig.ZulipAlarm.Interval) * time.Second

	queueMutex.Lock()
	defer queueMutex.Unlock()

	if len(zulipAlarmQueue) == 0 {
		return
	}

	newQueue := make([]string, 0, len(zulipAlarmQueue)) // to keep failed messages

	for _, msg := range zulipAlarmQueue {
		if err := sendZulipAlarm(msg); err != nil {
			newQueue = append(newQueue, msg)
		}

		time.Sleep(interval)
	}

	zulipAlarmQueue = newQueue
}

func sendZulipAlarm(message string) error {

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
