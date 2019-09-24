package ngerest

// Chat Trollbox Data
type Chat struct {
	ID        float32 `json:"id,omitempty"`
	Date      NGETime `json:"date"`
	User      string  `json:"user"`
	Message   string  `json:"message"`
	HTML      string  `json:"html"`
	FromBot   bool    `json:"fromBot,omitempty"`
	ChannelID float64 `json:"channelID,omitempty"`
}
