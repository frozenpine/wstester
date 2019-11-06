package ngerest

// AccessToken access token for auth
type AccessToken struct {
	ID      string   `json:"id"`
	TTL     float64  `json:"ttl,omitempty"`
	Created *NGETime `json:"created,omitempty"`
	UserID  float64  `json:"userId,omitempty"`
}
