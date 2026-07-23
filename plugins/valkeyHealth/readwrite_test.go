//go:build valkeyHealth

package main

import (
	"testing"

	lib "github.com/monobilisim/monokit2/lib"
)

func TestCheckValkeyReadWrite(t *testing.T) {
	err := lib.InitConfig(configFiles...)
	if err != nil {
		t.Fatalf("failed to init config: %v", err)
	}
	lib.InitializeDatabase()
	err = InitValkey()
	if err != nil {
		t.Fatalf("failed to init valkey: %v", err)
	}
	readable, writeable := TestValkeyReadWrite()
	if !readable {
		t.Errorf("expected readable, got false")
	}
	if !writeable {
		t.Errorf("expected writeable, got false")
	}
}
