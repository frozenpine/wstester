package models

import (
	"encoding/json"

	"github.com/frozenpine/ngerest"
)

// TradeResponse trade response structure
type TradeResponse struct {
	tableResponse

	Data []*ngerest.Trade `json:"data"`
}

// NewTradePartial make a new trade partial response
func NewTradePartial() *TradeResponse {
	partial := TradeResponse{}
	partial.Table = "trade"
	partial.Action = "partial"
	partial.Keys = []string{}
	partial.Types = make(map[string]string)
	partial.ForeignKeys = make(map[string]string)
	partial.Attributes = make(map[string]string)
	partial.Filter = make(map[string]string)

	return &partial
}

// String get structure's string format
func (td *TradeResponse) String() string {
	result, _ := json.Marshal(td)

	return string(result)
}

// Format format String output
func (td *TradeResponse) Format(format string) string {
	return td.String()
}
