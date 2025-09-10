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

// service = plugin name, module = specific module in the plugin, status = alarm status like "up" or "down"
// service, module name and status can be nil if not applicable
// instead of directly giving them as string, giving them as pointer to string
func SendZulipAlarm(message string, service *string, module *string, status *string) error {
	var lastAlarm ZulipAlarm
	var lastAlarms []ZulipAlarm
	var err error

	if service != nil {
		err = DB.
			Where("project_identifier = ? AND hostname = ? AND service = ? AND module = ?",
				GlobalConfig.ProjectIdentifier,
				GlobalConfig.Hostname,
				service,
				module).
			Order("id DESC").
			Limit(1).
			Find(&lastAlarm).Error

		err = DB.
			Where("project_identifier = ? AND hostname = ? AND service = ? AND module = ?",
				GlobalConfig.ProjectIdentifier,
				GlobalConfig.Hostname,
				service,
				module).
			Order("id DESC").
			Limit(GlobalConfig.ZulipAlarm.Limit).
			Find(&lastAlarms).Error
	}

	if service == nil {
		err = DB.
			Where("project_identifier = ? AND hostname = ?",
				GlobalConfig.ProjectIdentifier,
				GlobalConfig.Hostname).
			Order("id DESC").
			Limit(1).
			Find(&lastAlarm).Error

		err = DB.
			Where("project_identifier = ? AND hostname = ?",
				GlobalConfig.ProjectIdentifier,
				GlobalConfig.Hostname).
			Order("id DESC").
			Limit(GlobalConfig.ZulipAlarm.Limit).
			Find(&lastAlarms).Error
	}

	if err != nil {
		Logger.Error().Err(err).Msg("Failed to get last alarm from database")
		return err
	}

	// no extra checks needed if there are no previous alarms
	if len(lastAlarms) == 0 {
		return sendZulipAlarm(message)
	}

	// checking if alarms are duplicate
	firstStatus := lastAlarms[0].Status
	allSame := true
	for _, alarm := range lastAlarms {
		if alarm.Status != firstStatus {
			allSame = false
			break
		}
	}

	if status != nil && allSame && firstStatus == *status {
		var message string
		if service != nil {
			message = fmt.Sprintf("Zulip alarm limit (%s) has reached to limit for %s, skipping this one", GlobalConfig.ZulipAlarm.Limit, *service)
			Logger.Info().Str("service", *service).Msg(message)
		} else {
			message = fmt.Sprintf("Zulip alarm limit (%s) has reached to limit, skipping this one", GlobalConfig.ZulipAlarm.Limit)
			Logger.Info().Msg(message)
		}
		return fmt.Errorf(message)
	}

	if time.Since(lastAlarm.CreatedAt) < time.Duration(GlobalConfig.ZulipAlarm.Interval)*time.Minute {
		var message string
		if service != nil {
			message = fmt.Sprintf("Enough time is not passed since the last alarm from %s, skipping this one", *service)
			Logger.Info().Str("service", *service).Msg(message)
		} else {
			message = "Enough time is not passed since the last alarm, skipping this one"
			Logger.Info().Msg(message)
		}
		return fmt.Errorf(message)
	}

	return sendZulipAlarm(message)
}

func sendZulipAlarm(message string) error {

	if !GlobalConfig.ZulipAlarm.Enabled {
		Logger.Warn().Msg("Zulip alarm is bot enabled in the configuration")
		return fmt.Errorf("zulip alarm not enabled")
	}

	if GlobalConfig.ZulipAlarm.BotApi.Enabled {
		err := sendZulipBotApiAlarm(message)
		if err != nil {
			Logger.Error().Err(err).Msgf("failed to send Zulip Bot API alarm")
			return err
		}
	} else {
		for _, url := range GlobalConfig.ZulipAlarm.WebhookUrls {
			err := sendZulipWebhookAlarm(url, message)
			if err != nil {
				Logger.Error().Err(err).Msgf("Failed to send Zulip webhook alarm to %s", url)
				return err
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
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		Logger.Error().Msgf("Received non-2xx response: %d", resp.StatusCode)
		return fmt.Errorf("zulip webhook returned status %d", resp.StatusCode)
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
