package client

import (
	"fmt"
	"net/url"
	"strings"
)

type contextKey string

var (
	logLevel int

	// ContextAPIKey takes an APIKeyAuth as authentication for websocket
	ContextAPIKey = contextKey("apikey")

	symbolSubs = []string{"instrument", "orderBookL2", "trade", "order"}

	// PublicTopics public topics for subscribe without authentication
	PublicTopics = []string{"instrument", "orderBookL2", "orderBookL2_25", "trade", "quote"}
	// PrivateTopics private topics for subscribe must authenticated
	PrivateTopics = []string{"order", "execution", "position"}
)

// APIKeyAuth structure for api auth
type APIKeyAuth struct {
	Key     string
	Secret  string
	AuthURI string
}

// Config configuration for websocket
type Config struct {
	Symbol             string
	Scheme             string
	Host               string
	Port               int
	BaseURI            string
	HeartbeatInterval  int
	ReversHeartbeat    bool
	HeartbeatFailCount int
}

// ChangeHost change configuration's host
func (c *Config) ChangeHost(host string) error {
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
func (c *Config) GetURL() *url.URL {
	remote := url.URL{
		Scheme: c.Scheme,
		Host:   c.Host,
		Path:   c.BaseURI,
	}

	return &remote
}

// NewConfig to make a default new config
func NewConfig() *Config {
	cfg := Config{
		Symbol:             "XBTUSD",
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

// IsPublicTopic valid topic name in non case-sensitive
func IsPublicTopic(topic string) bool {
	for _, name := range PublicTopics {
		if strings.ToLower(topic) == strings.ToLower(name) {
			return true
		}
	}

	return false
}

// IsPrivateTopic valid topic name in non case-sensitive
func IsPrivateTopic(topic string) bool {
	for _, name := range PrivateTopics {
		if strings.ToLower(topic) == strings.ToLower(name) {
			return true
		}
	}

	return false
}

// IsValidTopic check topic name is valid in non case-sensitive
func IsValidTopic(topic string) bool {
	return IsPublicTopic(topic) || IsPrivateTopic(topic)
}
