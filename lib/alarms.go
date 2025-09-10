package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

var (
	zulipAlarmQueue []string
	queueMutex      sync.Mutex
)

// This function actually queues the messages to be sent for not overwhelm the Zulip server when there are many alerts at the same time.
func SendZulipAlarm(message string, service *string) bool {
	var lastAlarm ZulipAlarm
	var err error

	if service != nil {
		err = DB.
			Where("project_identifier = ? AND hostname = ? AND service = ?",
				GlobalConfig.ProjectIdentifier,
				GlobalConfig.Hostname,
				service).
			Order("id DESC").
			Limit(1).
			Find(&lastAlarm).Error
	}

	if service == nil {
		err = DB.
			Where("project_identifier = ? AND hostname = ?",
				GlobalConfig.ProjectIdentifier,
				GlobalConfig.Hostname).
			Order("id DESC").
			Limit(1).
			Find(&lastAlarm).Error
	}

	if err != nil {
		Logger.Error().Err(err).Msg("Failed to get last alarm from database")
		return false
	}

	if time.Since(lastAlarm.CreatedAt) < time.Duration(GlobalConfig.ZulipAlarm.Interval)*time.Minute {
		if service != nil {
			Logger.Info().Str("service", *service).Msgf("Enough time is not passed since the last alarm from %s, skipping this one", *service)
		} else {
			Logger.Info().Msgf("Enough time is not passed since the last alarm, skipping this one")
		}
		return false
	}

	queueMutex.Lock()
	zulipAlarmQueue = append(zulipAlarmQueue, message)
	queueMutex.Unlock()

	return true
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
	interval := 3 * time.Second

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
				continue
			}
			DB.Create(&ZulipAlarm{ProjectIdentifier: GlobalConfig.ProjectIdentifier, Hostname: GlobalConfig.Hostname, Content: message})
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

func CreateRedmineIssue(issue Issue) error {
	if !GlobalConfig.Redmine.Enabled {
		Logger.Warn().Msg("Redmine integration is not enabled in the configuration")
		return nil
	}

	var lastIssue Issue
	var err error

	if issue.Service != nil {
		err = DB.
			Where("project_identifier = ? AND hostname = ? AND service = ?",
				GlobalConfig.ProjectIdentifier,
				GlobalConfig.Hostname,
				issue.Service).
			Order("id DESC").
			Limit(1).
			Find(&lastIssue).Error
	}

	if issue.Service == nil {
		err = DB.
			Where("project_identifier = ? AND hostname = ?",
				GlobalConfig.ProjectIdentifier,
				GlobalConfig.Hostname).
			Order("id DESC").
			Limit(1).
			Find(&lastIssue).Error
	}

	if err != nil {
		Logger.Error().Err(err).Msg("Failed to get last alarm from database")
		return err
	}

	if time.Since(lastIssue.CreatedAt) < time.Duration(GlobalConfig.Redmine.Interval)*time.Minute {
		if issue.Service != nil {
			Logger.Info().Str("service", *issue.Service).Msgf("Enough time is not passed since the last Issue from %s, skipping this one", *issue.Service)
		} else {
			Logger.Info().Msgf("Enough time is not passed since the last alarm, skipping this one")
		}
	}

	if GlobalConfig.Redmine.ApiKey == "" || GlobalConfig.Redmine.Url == "" {
		Logger.Error().Msg("Redmine API key or URL not configured")
		return fmt.Errorf("redmine API key or URL not configured")
	}

	existingIssue := findRecentSimilarIssue(issue.Subject, 6)
	if existingIssue != nil {
		Logger.Info().Int("existing_issue_id", existingIssue.Id).Str("subject", issue.Subject).Msg("Found existing issue, reopening instead of creating new one")
		return reopenRedmineIssue(existingIssue.Id)
	}

	Logger.Info().Str("subject", issue.Subject).Msg("Creating new Redmine issue")
	return createNewRedmineIssue(issue)
}

func findRecentSimilarIssue(subject string, hoursBack int) *Issue {
	now := time.Now()
	hoursAgo := now.Add(-time.Duration(hoursBack) * time.Hour)

	var existingIssues []Issue
	result := DB.Where("subject = ? AND created_at > ?", subject, hoursAgo).Find(&existingIssues)
	if result.Error != nil {
		Logger.Error().Err(result.Error).Msg("Failed to query database for existing issues")
		return nil
	}

	if len(existingIssues) > 0 {
		Logger.Debug().Int("count", len(existingIssues)).Msg("Found existing issues in database")
		return &existingIssues[0]
	}

	return nil
}

func reopenRedmineIssue(issueId int) error {
	updateData := map[string]interface{}{
		"issue": map[string]interface{}{
			"status_id": 8,
		},
	}

	jsonBody, err := json.Marshal(updateData)
	if err != nil {
		Logger.Error().Err(err).Msg("Failed to marshal issue update request")
		return err
	}

	updateUrl := fmt.Sprintf("%s/issues/%d.json", GlobalConfig.Redmine.Url, issueId)
	req, err := http.NewRequest("PUT", updateUrl, bytes.NewBuffer(jsonBody))
	if err != nil {
		Logger.Error().Err(err).Str("url", updateUrl).Msg("Failed to create issue update request")
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Redmine-API-Key", GlobalConfig.Redmine.ApiKey)
	req.Header.Set("User-Agent", "Monokit/devel")

	client := &http.Client{Timeout: time.Second * 30}
	resp, err := client.Do(req)
	if err != nil {
		Logger.Error().Err(err).Str("url", updateUrl).Msg("Failed to send issue update request")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		Logger.Error().Int("status_code", resp.StatusCode).Str("response", string(body)).Msg("Failed to update Redmine issue")
		return fmt.Errorf("failed to update issue, status code: %d", resp.StatusCode)
	}

	Logger.Info().Int("issue_id", issueId).Msg("Successfully reopened Redmine issue")
	return nil
}

func createNewRedmineIssue(issue Issue) error {
	createData := map[string]interface{}{
		"issue": map[string]interface{}{
			"project_id":  GlobalConfig.ProjectIdentifier,
			"tracker_id":  issue.TrackerId,
			"subject":     issue.Subject,
			"description": issue.Description,
			"priority_id": issue.PriorityId,
		},
	}

	jsonBody, err := json.Marshal(createData)
	if err != nil {
		Logger.Error().Err(err).Msg("Failed to marshal issue creation request")
		return err
	}

	createUrl := GlobalConfig.Redmine.Url + "/issues.json"
	req, err := http.NewRequest("POST", createUrl, bytes.NewBuffer(jsonBody))
	if err != nil {
		Logger.Error().Err(err).Str("url", createUrl).Msg("Failed to create issue creation request")
		fmt.Println(err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Redmine-API-Key", GlobalConfig.Redmine.ApiKey)
	req.Header.Set("User-Agent", "Monokit/devel")

	client := &http.Client{Timeout: time.Second * 30}
	resp, err := client.Do(req)
	if err != nil {
		Logger.Error().Err(err).Str("url", createUrl).Msg("Failed to send issue creation request")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		Logger.Error().Int("status_code", resp.StatusCode).Str("response", string(body)).Msg("Failed to create Redmine issue")
		return fmt.Errorf("failed to create issue, status code: %d", resp.StatusCode)
	}

	var response struct {
		Issue struct {
			Id int `json:"id"`
		} `json:"issue"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		Logger.Error().Err(err).Msg("Failed to read response body")
		return err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		Logger.Error().Err(err).Msg("Failed to parse issue creation response")
		return err
	}

	issue.Id = response.Issue.Id
	result := DB.Create(&issue)
	if result.Error != nil {
		Logger.Error().Err(result.Error).Int("issue_id", issue.Id).Msg("Failed to store issue in local database")
	}

	Logger.Info().Int("issue_id", response.Issue.Id).Str("subject", issue.Subject).Msg("Successfully created new Redmine issue")
	return nil
}
