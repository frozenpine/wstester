package server

import (
	"crypto/md5"
	"fmt"
	"net"
	"strconv"
	"strings"

	uuid "github.com/satori/go.uuid"
)

// SvrConfig websocket listen config
type SvrConfig struct {
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
func (c *SvrConfig) ChangeListen(addr string) error {
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
func (c *SvrConfig) GetListenAddr() string {
	return fmt.Sprintf("%s:%d", c.Listen.String(), c.Port)
}

// GetNS get server's namespace from listen addr
func (c *SvrConfig) GetNS() uuid.UUID {
	nsString := fmt.Sprintf("%v:%s", c.FrontID, c.GetListenAddr())
	nsHash := md5.Sum([]byte(nsString))

	return uuid.Must(uuid.FromBytes(nsHash[:]))
}

// NewSvrConfig create a new server config
func NewSvrConfig() *SvrConfig {
	cfg := SvrConfig{
		Listen:       net.ParseIP("0.0.0.0"),
		Port:         9988,
		BaseURI:      "/realtime",
		SignatureURI: "/api/v1/signature",

		WelcomMsg: "Welcome to the BTCMEX Realtime API.",
		DocsURI:   "https://docs.btcmex.com",
		FrontID:   "0",

		ConnectLimit: 40,

		HeartbeatInterval:  15,
		ReversHeartbeat:    false,
		HeartbeatFailCount: 3,
	}

	return &cfg
}
