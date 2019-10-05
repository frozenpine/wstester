package models

import (
	"encoding/json"

	"github.com/frozenpine/ngerest"
)

// TradeResponse trade response structure
type TradeResponse struct {
	tableResponse

	Data []ngerest.Trade `json:"data"`
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
