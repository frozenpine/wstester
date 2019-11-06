package ngerest

// Insurance Fund Data
type Insurance struct {
	Currency      string   `json:"currency"`
	Timestamp     *NGETime `json:"timestamp"`
	WalletBalance float32  `json:"walletBalance,omitempty"`
}
