package models

import "github.com/frozenpine/ngerest"

// MBLResponse trade response structure
type MBLResponse struct {
	tableResponse

	Data []ngerest.OrderBookL2 `json:"data"`
}
