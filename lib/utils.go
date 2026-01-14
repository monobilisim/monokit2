package lib

import (
	"os"
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
