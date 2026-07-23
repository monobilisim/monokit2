//go:build redisHealth

package main

import (
	"testing"
)

func TestDetectRedis(t *testing.T) {
	if !DetectRedis() {
		t.Errorf("expected redis to be detected")
	}
}
func TestIsRedisSentinel(t *testing.T) {
	setupRedisTest(t)
	isSentinel := IsRedisSentinel()
	t.Logf("isSentinel: %v", isSentinel)
}
func TestCheckRedisConnection(t *testing.T) {
	setupRedisTest(t)
	err := CheckRedisConnection()
	if err != nil {
		t.Fatalf("expected redis connection to succeed, got error: %v", err)
	}
}
func TestCheckRedisSlaveCount(t *testing.T) {
	setupRedisTest(t)
	isMaster, count, isSentinel, slaveOK := CheckRedisSlaveCount()
	t.Logf("isMaster: %v, count: %v, isSentinel: %v, slaveOK: %v", isMaster, count, isSentinel, slaveOK)

}
