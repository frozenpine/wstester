package server

// MBLCache retrive & store mbl data
type MBLCache struct {
	rspChannel
}

// NewMBLCache make a new MBL cache.
func NewMBLCache() *MBLCache {
	mbl := MBLCache{}

	return &mbl
}
