package ngerest

// Funding Swap Funding History
type Funding struct {
	Timestamp        NGETime `json:"timestamp"`
	Symbol           string  `json:"symbol"`
	FundingInterval  NGETime `json:"fundingInterval,omitempty"`
	FundingRate      float64 `json:"fundingRate,omitempty"`
	FundingRateDaily float64 `json:"fundingRateDaily,omitempty"`
}
