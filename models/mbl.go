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

// ToString get structure's string format
func (mbl *MBLResponse) ToString() string {
	result, _ := json.Marshal(mbl)

	return string(result)
}

// Format format ToString output
func (mbl *MBLResponse) Format(format string) string {
	return mbl.ToString()
}
