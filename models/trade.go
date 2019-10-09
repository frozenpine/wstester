package models

import (
	"github.com/frozenpine/ngerest"
)

// TradeResponse trade response structure
type TradeResponse struct {
	tableResponse

	Data []*ngerest.Trade `json:"data"`
}

// NewTradePartial make a new trade partial response
func NewTradePartial() *TradeResponse {
	partial := TradeResponse{}
	partial.Table = "trade"
	partial.Action = "partial"
	partial.Keys = []string{}
	partial.Types = make(map[string]string)
	partial.ForeignKeys = make(map[string]string)
	partial.Attributes = make(map[string]string)
	partial.Filter = make(map[string]string)

	return &partial
}
