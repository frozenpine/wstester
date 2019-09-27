package models

import (
	"encoding/json"

	"github.com/frozenpine/ngerest"
)

// Response common functions for response
type Response interface {
	ToString() string
	Format(string) string
}

// OperationRequest request to websocket
type OperationRequest struct {
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

// ToString get structure's string format
func (info *InfoResponse) ToString() string {
	result, _ := json.Marshal(info)

	return string(result)
}

// Format format ToString output
func (info *InfoResponse) Format(format string) string {
	return info.ToString()
}

// AuthResponse authentication response
type AuthResponse struct {
	Success bool                   `json:"success"`
	Request map[string]interface{} `json:"request"`
}

// ToString get structure's string format
func (auth *AuthResponse) ToString() string {
	result, _ := json.Marshal(auth)

	return string(result)
}

// Format format ToString output
func (auth *AuthResponse) Format(format string) string {
	return auth.ToString()
}

// SubscribeResponse subscribe response
type SubscribeResponse struct {
	Success   bool             `json:"success"`
	Subscribe string           `json:"subscribe"`
	Request   OperationRequest `json:"request"`
}

// ToString get structure's string format
func (sub *SubscribeResponse) ToString() string {
	result, _ := json.Marshal(sub)

	return string(result)
}

// Format format ToString output
func (sub *SubscribeResponse) Format(format string) string {
	return sub.ToString()
}

type tableResponse struct {
	Table  string `json:"table"`
	Action string `json:"action"`

	Keys        []string          `json:"keys,omitempty"`
	Types       map[string]string `json:"types,omitempty"`
	ForeignKeys map[string]string `json:"foreignKeys,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	Filter      map[string]string `json:"filter,omitempty"`
}
