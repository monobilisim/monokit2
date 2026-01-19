package lib

import (
	"database/sql/driver"
	"fmt"
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
