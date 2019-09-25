package modules

import (
	"fmt"
	"net/url"
	"strings"
)

var (
	logLevel int
)

// WsConfig configuration for websocket
type WsConfig struct {
	Scheme             string
	Host               string
	Port               int
	BaseURI            string
	HeartbeatInterval  int
	ReversHeartbeat    bool
	HeartbeatFailCount int
}

// ChangeHost change configuration's host
func (c *WsConfig) ChangeHost(host string) error {
	result, err := url.Parse(host)

	if err != nil {
		return err
	}

	switch {
	case strings.Contains(result.Scheme, "http"):
		c.Scheme = strings.Replace(result.Scheme, "http", "ws", 1)
	case strings.Contains(result.Scheme, "ws"):
		c.Scheme = result.Scheme
	default:
		return fmt.Errorf("invalid scheme: %s", result.Scheme)
	}

	c.Host = result.Host
	c.BaseURI = result.Path

	return nil
}

// GetURL to convert confiuration to URL instance
func (c *WsConfig) GetURL() *url.URL {
	remote := url.URL{
		Scheme: c.Scheme,
		Host:   c.Host,
		Path:   c.BaseURI,
	}

	return &remote
}

// NewConfig to make a default new config
func NewConfig() *WsConfig {
	cfg := WsConfig{
		Scheme:             "wss",
		Host:               "www.btcmex.com",
		BaseURI:            "/realtime",
		HeartbeatInterval:  15,
		ReversHeartbeat:    false,
		HeartbeatFailCount: 3,
	}

	return &cfg
}

// SetLogLevel set log level to display more detailed log info
func SetLogLevel(lvl int) {
	logLevel = lvl
}
