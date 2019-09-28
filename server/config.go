package server

import "fmt"

// WsConfig websocket listen config
type WsConfig struct {
	Listen  string
	Port    int
	BaseURI string
}

// GetListenAddr get configured listen addr for server
func (c *WsConfig) GetListenAddr() string {
	return fmt.Sprintf("%s:%d", c.Listen, c.Port)
}
