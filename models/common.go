package models

import (
	"github.com/frozenpine/ngerest"
)

// SubscribeRequest request to websocket
type SubscribeRequest struct {
	Operation string   `json:"op"`
	Args      []string `json:"args"`
}

// InfoResponse welcome message
type InfoResponse struct {
	Info      string                 `json:"info"`
	Version   string                 `json:"version"`
	Timestamp ngerest.NGETime        `json:"timestamp"`
	Docs      string                 `json:"docs"`
	Limit     map[string]interface{} `json:"limit"`
	FrontID   string                 `json:"frontId"`
	SessionID string                 `json:"sessionId"`
}

// SubscribeResponse subscribe response
type SubscribeResponse struct {
	Success   bool             `json:"success"`
	Subscribe string           `json:"subscribe"`
	Request   SubscribeRequest `json:"request"`
}
