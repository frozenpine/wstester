package ngerest

// OrderBookL2 level 2 orderbook
type OrderBookL2 struct {
	Symbol string  `json:"symbol"`
	ID     float32 `json:"id"`
	Side   string  `json:"side"`
	Size   float32 `json:"size,omitempty"`
	Price  float64 `json:"price,omitempty"`
}
