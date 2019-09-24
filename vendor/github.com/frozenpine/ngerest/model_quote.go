package ngerest

// Quote Best Bid/Offer Snapshots & Historical Bins
type Quote struct {
	Timestamp NGETime `json:"timestamp"`
	Symbol    string  `json:"symbol"`
	BidSize   float32 `json:"bidSize,omitempty"`
	BidPrice  float64 `json:"bidPrice,omitempty"`
	AskPrice  float64 `json:"askPrice,omitempty"`
	AskSize   float32 `json:"askSize,omitempty"`
}
