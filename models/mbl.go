package models

import "github.com/frozenpine/ngerest"

// MBLResponse trade response structure
type MBLResponse struct {
	Table  string                `json:"table"`
	Action string                `json:"action"`
	Data   []ngerest.OrderBookL2 `json:"data"`

	Keys        []string          `json:"keys,omitempty"`
	Types       map[string]string `json:"types,omitempty"`
	ForeignKeys map[string]string `json:"foreignKeys,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	Filter      map[string]string `json:"filter,omitempty"`
}
