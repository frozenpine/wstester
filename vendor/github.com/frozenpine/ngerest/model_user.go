package ngerest

// User Account Operations
type User struct {
	ID           float32          `json:"id,omitempty"`
	OwnerID      float32          `json:"ownerId,omitempty"`
	Firstname    string           `json:"firstname,omitempty"`
	Lastname     string           `json:"lastname,omitempty"`
	Username     string           `json:"username"`
	Email        string           `json:"email"`
	Phone        string           `json:"phone,omitempty"`
	Created      *NGETime         `json:"created,omitempty"`
	LastUpdated  *NGETime         `json:"lastUpdated,omitempty"`
	Preferences  *UserPreferences `json:"preferences,omitempty"`
	TFAEnabled   string           `json:"TFAEnabled,omitempty"`
	AffiliateID  string           `json:"affiliateID,omitempty"`
	PgpPubKey    string           `json:"pgpPubKey,omitempty"`
	Country      string           `json:"country,omitempty"`
	GeoipCountry string           `json:"geoipCountry,omitempty"`
	GeoipRegion  string           `json:"geoipRegion,omitempty"`
	Typ          string           `json:"typ,omitempty"`
}

// UserDefaultAPIKey user's default api key
type UserDefaultAPIKey struct {
	UserID    string `json:"userId"`
	APIKey    string `json:"apiKey"`
	APISecret string `json:"secret"`
}
