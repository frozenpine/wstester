package server

// InstrumentCache retrive & store instrument data
type InstrumentCache struct {
	channel
}

// NewInstrumentCache make a new instrument cache.
func NewInstrumentCache() *InstrumentCache {
	ins := InstrumentCache{}

	return &ins
}
