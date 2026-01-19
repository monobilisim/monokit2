package lib

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func IsTestMode() bool {
	v, ok := os.LookupEnv("TEST")
	return ok && (v == "1" || v == "true" || v == "yes")
}

func GetLastCronInterval(name string) CronInterval {
	var ci CronInterval
	result := DB.Where(&CronInterval{Name: name}).First(&ci)

	if result.RowsAffected == 0 {
		newCi := CronInterval{
			Name:    name,
			LastRun: nil,
		}
		DB.Create(&newCi)
		return newCi
	}

	return ci
}

func CreateOrUpdateCronInterval(name string) {
	now := time.Now()
	ci := GetLastCronInterval(name)
	ci.LastRun = &now
	DB.Save(&ci)
}

// Convert redmine date formatting to golang and database compatible format
func (d *RedmineDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "" || s == "null" {
		d.Time = time.Time{}
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		t, err = time.Parse(time.RFC3339, s)
	}
	d.Time = t
	return err
}

func (d RedmineDate) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", d.Time.Format("2006-01-02"))), nil
}

func (d RedmineDate) Value() (driver.Value, error) {
	return d.Time, nil
}

func (d *RedmineDate) Scan(value interface{}) error {
	if value == nil {
		d.Time = time.Time{}
		return nil
	}
	if t, ok := value.(time.Time); ok {
		d.Time = t
		return nil
	}
	return fmt.Errorf("cannot scan %v into RedmineDate", value)
}

func (r RedmineAPIObject) Value() (driver.Value, error) {
	return json.Marshal(r)
}

func (r *RedmineAPIObject) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return errors.New("type assertion to []byte or string failed")
	}
	return json.Unmarshal(b, &r)
}

func GetIssueFromRedmine(id int) (RedmineIssue, error) {
	var issue RedmineIssue

	url := fmt.Sprintf("%s/issues/%d.json", GlobalConfig.Redmine.Url, id)
	reqGet, err := http.NewRequest("GET", url, nil)
	if err != nil {
		Logger.Error().Err(err).Str("url", url).Msg("Failed to create issue fetch request")
		return issue, err
	}

	reqGet.Header.Set("Content-Type", "application/json")
	reqGet.Header.Set("X-Redmine-API-Key", GlobalConfig.Redmine.ApiKey)
	reqGet.Header.Set("User-Agent", "Monokit/devel")

	getClient := &http.Client{Timeout: time.Second * 30}
	resp, err := getClient.Do(reqGet)
	if err != nil {
		Logger.Error().Err(err).Str("url", url).Msg("Failed to send issue fetch request")
		return issue, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		Logger.Error().Int("status_code", resp.StatusCode).Str("response", string(body)).Msg("Failed to fetch Redmine issue")
		return issue, fmt.Errorf("failed to fetch issue, status code: %d", resp.StatusCode)
	}

	bodyGet, err := io.ReadAll(resp.Body)
	if err != nil {
		Logger.Error().Err(err).Msg("Failed to read fetch response body")
		return issue, err
	}

	err = json.Unmarshal(bodyGet, &issue)
	if err != nil {
		Logger.Error().Err(err).Msg("Failed to parse fetch response")
		return issue, err
	}

	return issue, nil
}
