package models

import (
	"encoding/json"

	"github.com/frozenpine/ngerest"
)

// MBLResponse trade response structure
type MBLResponse struct {
	tableResponse

	Data []*ngerest.OrderBookL2 `json:"data"`
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

// GetAction get action for response
func (mbl *MBLResponse) GetAction() string {
	return mbl.Action
}

// GetData get data for reponse
func (mbl *MBLResponse) GetData() []interface{} {
	var data []interface{}

	for _, d := range mbl.Data {
		data = append(data, d)
	}

	return data
}
