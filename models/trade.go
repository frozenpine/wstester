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

// ToString get structure's string format
func (td *TradeResponse) ToString() string {
	result, _ := json.Marshal(td)

	return string(result)
}

// Format format ToString output
func (td *TradeResponse) Format(format string) string {
	return td.ToString()
}
