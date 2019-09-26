package models

import "github.com/frozenpine/ngerest"

// TradeResponse trade response structure
type TradeResponse struct {
	tableResponse

	Data []ngerest.Trade `json:"data"`
}
