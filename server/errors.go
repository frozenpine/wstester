package server

// ErrAPIExpires api signature expires error
type ErrAPIExpires struct {
	expires int64
}

func (e *ErrAPIExpires) Error() string {
	return ""
}

// NewAPIExpires create api signature expires error
func NewAPIExpires(expires int64) *ErrAPIExpires {
	err := ErrAPIExpires{
		expires: expires,
	}

	return &err
}
