package server

// TradeCache retrive & store trade data
type TradeCache struct {
	channel
}

// NewTradeCache make a new trade cache.
func NewTradeCache() *TradeCache {
	td := TradeCache{}

	return &td
}
