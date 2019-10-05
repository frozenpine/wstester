package models

import (
	"encoding/json"

	"github.com/frozenpine/ngerest"
)

// MBLResponse trade response structure
type MBLResponse struct {
	tableResponse

	Data []ngerest.OrderBookL2 `json:"data"`
}

// String get structure's string format
func (mbl *MBLResponse) String() string {
	result, _ := json.Marshal(mbl)

	return string(result)
}

// Format format String output
func (mbl *MBLResponse) Format(format string) string {
	return mbl.String()
}
