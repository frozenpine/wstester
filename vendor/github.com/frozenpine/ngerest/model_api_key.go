package ngerest

// APIKeyInfo persistent API Keys for Developers
type APIKeyInfo struct {
	ID          string   `json:"id"`
	Secret      string   `json:"secret"`
	Name        string   `json:"name"`
	Nonce       float32  `json:"nonce"`
	Cidr        string   `json:"cidr,omitempty"`
	Permissions []XAny   `json:"permissions,omitempty"`
	Enabled     bool     `json:"enabled,omitempty"`
	UserID      float32  `json:"userId"`
	Created     *NGETime `json:"created,omitempty"`
}
