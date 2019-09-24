package ngerest

// IndexComposite index composite
type IndexComposite struct {
	Timestamp   NGETime `json:"timestamp"`
	Symbol      string  `json:"symbol,omitempty"`
	IndexSymbol string  `json:"indexSymbol,omitempty"`
	Reference   string  `json:"reference,omitempty"`
	LastPrice   float64 `json:"lastPrice,omitempty"`
	Weight      float64 `json:"weight,omitempty"`
	Logged      NGETime `json:"logged,omitempty"`
}
