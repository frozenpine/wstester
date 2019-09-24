package models

import (
	"time"
)

type hbValue int8

const (
	ping hbValue = 1
	pong hbValue = -1
)

// HeartBeat message
type HeartBeat struct {
	value hbValue
	ts    time.Time
}

// IsPing to test whether heartbeat is ping type
func (hb *HeartBeat) IsPing() bool {
	return hb.value == ping
}

// IsPong to test whether heartbeat is pong type
func (hb *HeartBeat) IsPong() bool {
	return hb.value == pong
}

// GetTime to get heartbeat message's timestamp
func (hb *HeartBeat) GetTime() time.Time {
	return hb.ts
}

// Value to get heartbeat value
func (hb *HeartBeat) Value() int {
	return int(hb.value)
}

// Type to get heartbeat type in string
func (hb *HeartBeat) Type() string {
	if hb.value == ping {
		return "Ping"
	}

	if hb.value == pong {
		return "Pong"
	}

	return "Invalid type."
}

// ToString to get heartbeat message in string format
func (hb *HeartBeat) ToString() string {
	return hb.Type() + ": " + hb.ts.Format(time.RFC1123)
}

// NewPing create a heatbeat message in ping type
func NewPing() *HeartBeat {
	return &HeartBeat{value: ping, ts: time.Now()}
}

// NewPong create a heatbeat message in pong type
func NewPong() *HeartBeat {
	return &HeartBeat{value: pong, ts: time.Now()}
}
