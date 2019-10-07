package server

// MBLCache retrive & store mbl data
type MBLCache struct {
	channel
}

// NewMBLCache make a new MBL cache.
func NewMBLCache() *MBLCache {
	mbl := MBLCache{}

	return &mbl
}
