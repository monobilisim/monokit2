//go:build redisHealth

package main

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	lib "github.com/monobilisim/monokit2/lib"
	"github.com/redis/go-redis/v9"
)

var (
	rdb      *redis.Client
	ctx      context.Context
	lastRole string
)

func isServiceActive(service string) bool {
	out, err := exec.Command("systemctl", "is-active", service).Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "active"
}
func DetectRedis() bool {
	if !isServiceActive("redis.service") &&
		!isServiceActive("redis-server.service") {
		return false
	}

	client := redis.NewClient(&redis.Options{
		Addr:       "localhost:6379",
		Password:   "",
		DB:         0,
		MaxRetries: 1,
	})
	defer client.Close()
	ctx = context.Background()
	pong, err := client.Ping(ctx).Result()
	if err != nil {
		return false
	}
	return pong == "PONG"
}
func InitRedis() error {
	port := lib.DBConfig.Redis.Credentials.Port
	if port == "" {
		port = "6379"
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:       fmt.Sprintf("localhost:%s", port),
		Password:   lib.DBConfig.Redis.Credentials.Password,
		DB:         0,
		MaxRetries: 5,
	})

	ctx = context.Background()
	ping, err := rdb.Ping(ctx).Result()
	if err != nil {
		return err
	}
	if ping != "PONG" {
		return fmt.Errorf("redis ping failed")
	}
	return nil
}
func TestRedisReadWrite() (bool, bool) {
	if rdb == nil {
		return false, false
	}
	err := rdb.Set(ctx, "monokit_health_test", "ok", 0).Err()
	if err != nil {
		return false, false
	}
	val, err := rdb.Get(ctx, "monokit_health_test").Result()
	if err != nil {
		return false, false
	}
	_ = rdb.Del(ctx, "monokit_health_test").Err()
	if val != "ok" {
		return false, false
	}
	return true, true
}
func IsRedisMaster() bool {
	if rdb == nil {
		return false
	}
	info, err := rdb.Info(ctx, "Replication").Result()
	if err != nil {
		return false
	}

	lines := strings.Split(info, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "role:master" {
			return true
		}
		if line == "role:slave" {
			return false
		}
	}
	return false
}
func GetActualSlaveCount() int {
	if rdb == nil {
		return 0
	}
	info, err := rdb.Info(ctx, "Replication").Result()
	if err != nil {
		return 0
	}
	lines := strings.Split(info, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "connected_slaves:") {
			countStr := strings.TrimPrefix(line, "connected_slaves:")
			countStr = strings.TrimSpace(countStr)
			count, err := strconv.Atoi(countStr)
			if err == nil {
				return count
			}
		}
	}
	return 0
}
func IsRedisSentinel() bool {
	conn, err := net.DialTimeout("tcp", "localhost:26379", time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	client := redis.NewClient(&redis.Options{
		Addr:       "localhost:26379",
		Password:   "",
		DB:         0,
		MaxRetries: 3,
	})
	defer client.Close()
	sentinelCtx := context.Background()
	info, err := client.Info(sentinelCtx, "Server").Result()
	if err != nil {
		return false
	}
	lines := strings.Split(info, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "redis_mode:sentinel" {
			return true
		}
	}
	return false
}
func CheckSlaveCount(expected int) bool {
	if !IsRedisSentinel() {
		return true
	}

	if !IsRedisMaster() {
		return true
	}

	actual := GetActualSlaveCount()

	return actual == expected
}
func GetRedisVersion() string {
	if rdb == nil {
		return "Unknown"
	}
	info, err := rdb.Info(ctx, "Server").Result()
	if err != nil {
		return "Unknown"
	}
	lines := strings.Split(info, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "redis_version:") {
			return strings.TrimPrefix(line, "redis_version:")
		}
	}
	return "Unknown"
}
func GetConnectedClients() int {
	if rdb == nil {
		return 0
	}
	info, err := rdb.Info(ctx, "Clients").Result()
	if err != nil {
		return 0
	}
	lines := strings.Split(info, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "connected_clients:") {
			value := strings.TrimPrefix(line, "connected_clients:")
			value = strings.TrimSpace(value)

			count, err := strconv.Atoi(value)
			if err == nil {
				return count
			}
		}
	}
	return 0
}
func GetRedisUptime() int {
	if rdb == nil {
		return 0
	}
	info, err := rdb.Info(ctx, "Server").Result()
	if err != nil {
		return 0
	}
	lines := strings.Split(info, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "uptime_in_seconds:") {
			value := strings.TrimPrefix(line, "uptime_in_seconds:")
			value = strings.TrimSpace(value)

			seconds, err := strconv.Atoi(value)
			if err == nil {
				return seconds
			}
		}
	}
	return 0
}
func FormatUptime(seconds int) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60

	return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
}
func GetUsedMemory() string {
	if rdb == nil {
		return "Unknown"
	}
	info, err := rdb.Info(ctx, "Memory").Result()
	if err != nil {
		return "Unknown"
	}
	lines := strings.Split(info, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "used_memory_human:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "used_memory_human:"))
		}
	}
	return "Unknown"
}
func CheckRoleChange(currentRole string) bool {
	if lastRole == "" {
		lastRole = currentRole
		return false
	}
	if lastRole != currentRole {
		lastRole = currentRole
		return true
	}
	return false
}
func GetPersistenceMode() string {
	if rdb == nil {
		return "Unknown"
	}
	info, err := rdb.Info(ctx, "Persistence").Result()
	if err != nil {
		return "Unknown"
	}
	var rdbEnabled bool
	var aofEnabled bool
	lines := strings.Split(info, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "rdb_bgsave_in_progress:1" ||
			line == "rdb_last_bgsave_status:ok" {
			rdbEnabled = true
		}
		if line == "aof_enabled:1" {
			aofEnabled = true
		}
	}
	switch {
	case rdbEnabled && aofEnabled:
		return "RDB + AOF"
	case rdbEnabled:
		return "RDB"
	case aofEnabled:
		return "AOF"
	default:
		return "Disabled"
	}
}
