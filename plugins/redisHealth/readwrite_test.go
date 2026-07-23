//go:build redisHealth

package main

import (
	"testing"

	lib "github.com/monobilisim/monokit2/lib"
)

func TestCheckRedisReadWrite(t *testing.T) {
	err := lib.InitConfig(configFiles...)
	if err != nil {
		t.Fatalf("failed to init config: %v", err)
	}
	lib.InitializeDatabase()
	err = InitRedis()
	if err != nil {
		t.Fatalf("failed to init redis: %v", err)
	}
	readable, writeable := TestRedisReadWrite()
	if !readable {
		t.Errorf("expected readable, got false")
	}
	if !writeable {
		t.Errorf("expected writeable, got false")
	}
}
