package server

import "fmt"

// WsConfig websocket listen config
type WsConfig struct {
	Listen             string
	Port               int
	BaseURI            string
	SignatureURI       string
	WelcomMsg          string
	DocsURI            string
	FrontID            string
	HeartbeatInterval  int
	ReversHeartbeat    bool
	HeartbeatFailCount int
}

// GetListenAddr get configured listen addr for server
func (c *WsConfig) GetListenAddr() string {
	return fmt.Sprintf("%s:%d", c.Listen, c.Port)
}

// NewConfig create a new server config
func NewConfig() *WsConfig {
	cfg := WsConfig{
		Listen:             "0.0.0.0",
		Port:               9988,
		BaseURI:            "/realtime",
		SignatureURI:       "/api/v1/signature",
		HeartbeatInterval:  15,
		ReversHeartbeat:    false,
		HeartbeatFailCount: 3,
	}

	return &cfg
}
