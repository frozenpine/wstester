package models

import "github.com/frozenpine/ngerest"

// InstrumentResponse instrument response structure
type InstrumentResponse struct {
	tableResponse

	Data []ngerest.Instrument `json:"data"`
}
