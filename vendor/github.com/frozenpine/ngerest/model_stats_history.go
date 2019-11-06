package ngerest

// StatsHistory history states
type StatsHistory struct {
	Date       *NGETime `json:"date"`
	RootSymbol string   `json:"rootSymbol"`
	Currency   string   `json:"currency,omitempty"`
	Volume     float32  `json:"volume,omitempty"`
	Turnover   float32  `json:"turnover,omitempty"`
}
