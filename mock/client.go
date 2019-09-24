package mock

import (
	"context"
	"log"

	"github.com/gorilla/websocket"
)

// Client client instance
type Client struct {
	cfg       *WsConfig
	ws        *websocket.Conn
	connected bool
	ctx       context.Context
}

// Host to get remote host string
func (c *Client) Host() string {
	return c.cfg.GetURL().String()
}

// IsConnected to specify if client is connected to remote host
func (c *Client) IsConnected() bool {
	return c.connected && c.ws != nil
}

// Connect to remote host
func (c *Client) Connect(ctx context.Context) error {
	remote := c.cfg.GetURL()

	remote.RawQuery = "subscribe=instrument:XBTUSD,orderBookL2:XBTUSD,trade:XBTUSD"

	log.Println("Connect to:", remote.String())

	conn, rsp, err := websocket.DefaultDialer.DialContext(
		ctx, remote.String(), nil)

	if err != nil {
		log.Printf("Fail to connect[%s]: %v\n%s",
			remote.String(), err, rsp.Status)

		return err
	}

	c.ws = conn
	c.connected = true

	return nil
}

// NewClient create a new mock client instance
func NewClient(cfg *WsConfig) {
}

func heartbeatHandler(ctx context.Context, ws *websocket.Conn) {

}
