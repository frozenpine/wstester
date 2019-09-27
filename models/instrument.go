package models

import (
	"encoding/json"

	"github.com/frozenpine/ngerest"
)

// InstrumentResponse instrument response structure
type InstrumentResponse struct {
	tableResponse

	Data []ngerest.Instrument `json:"data"`
}

// ToString get structure's string format
func (ins *InstrumentResponse) ToString() string {
	result, _ := json.Marshal(ins)

	return string(result)
}

// Format format ToString output
func (ins *InstrumentResponse) Format(format string) string {
	return ins.ToString()
}
