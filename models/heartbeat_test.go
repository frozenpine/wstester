package models

import (
	"testing"
)

func HeartBeatTest(t *testing.T) {
	hb := NewHeartbeat(false)

	if !hb.IsPing() {
		t.Fatal("invalid heartbeat ping.")
	}
}
