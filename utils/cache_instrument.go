package utils

// InstrumentCache retrive & store instrument data
type InstrumentCache struct {
	rspChannel
}

// NewInstrumentCache make a new instrument cache.
func NewInstrumentCache() *InstrumentCache {
	ins := InstrumentCache{}

	return &ins
}
