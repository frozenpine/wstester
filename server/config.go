package server

import (
	"crypto/md5"
	"fmt"
	"net"
	"strconv"
	"strings"

	uuid "github.com/satori/go.uuid"
)

// SvrContextKey context key for server
type SvrContextKey string

const (
	defaultListen       = "0.0.0.0"
	defaultPort         = 9988
	defaultBaseURI      = "/realtime"
	defaultSignatureURI = "/api/v1/signature"
	defaultWelcomMsg    = "Welcome to the BTCMEX Realtime API."
	defaultDocURI       = "https://docs.btcmex.com"
	defaultID           = "0"
	defaultHBInterval   = 15
	defaultHBFail       = 3
	isReverseHB         = false

	// SvrConfigKey context key for SvrConfig
	SvrConfigKey = SvrContextKey("config")
)

// Config websocket listen config
type Config struct {
	Listen       net.IP
	Port         int
	BaseURI      string
	SignatureURI string

	WelcomMsg string
	DocsURI   string
	FrontID   string

	ConnectLimit int

	HeartbeatInterval  int
	ReversHeartbeat    bool
	HeartbeatFailCount int
}

// ChangeListen change server listen address
func (c *Config) ChangeListen(addr string) error {
	errInvalidAddr := fmt.Errorf("invalid addr: %s", addr)
	addrTuple := strings.Split(addr, ":")
	if len(addrTuple) != 2 {
		return errInvalidAddr
	}

	if ip := net.ParseIP(addrTuple[0]); ip != nil {
		c.Listen = ip
	} else {
		return errInvalidAddr
	}

	port, err := strconv.Atoi(addrTuple[1])

	if err != nil {
		return err
	}
	c.Port = port

	return nil
}

// GetListenAddr get configured listen addr for server
func (c *Config) GetListenAddr() string {
	return fmt.Sprintf("%s:%d", c.Listen.String(), c.Port)
}

// GetNS get server's namespace from listen addr
func (c *Config) GetNS() uuid.UUID {
	nsString := fmt.Sprintf("%v:%s", c.FrontID, c.GetListenAddr())
	nsHash := md5.Sum([]byte(nsString))

	return uuid.Must(uuid.FromBytes(nsHash[:]))
}

// NewConfig create a new server config
func NewConfig() *Config {
	cfg := Config{
		Listen:       net.ParseIP(defaultListen),
		Port:         defaultPort,
		BaseURI:      defaultBaseURI,
		SignatureURI: defaultSignatureURI,

		WelcomMsg: defaultWelcomMsg,
		DocsURI:   defaultDocURI,
		FrontID:   defaultID,

		ConnectLimit: 40,

		HeartbeatInterval:  defaultHBInterval,
		ReversHeartbeat:    isReverseHB,
		HeartbeatFailCount: defaultHBFail,
	}

	return &cfg
}
