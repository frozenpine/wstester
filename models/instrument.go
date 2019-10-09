package models

import (
	"encoding/json"

	"github.com/frozenpine/ngerest"
)

// InstrumentResponse instrument response structure
type InstrumentResponse struct {
	tableResponse

	Data []*ngerest.Instrument `json:"data"`
}

// String get structure's string format
func (ins *InstrumentResponse) String() string {
	result, _ := json.Marshal(ins)

	return string(result)
}

// Format format String output
func (ins *InstrumentResponse) Format(format string) string {
	return ins.String()
}
