package modules

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/frozenpine/wstester/models"
	"github.com/gorilla/websocket"
)

var (
	pingPattern       = []byte("ping")
	pongPattern       = []byte("pong")
	infoPattern       = []byte(`"info"`)
	instrumentPattern = []byte(`"instrument"`)
	tradePattern      = []byte(`"trade"`)
	mblPattern        = []byte(`"orderBook`)
	subPattern        = []byte(`"subscribe"`)
)

// Client client instance
type Client struct {
	cfg       *WsConfig
	ws        *websocket.Conn
	connected bool
	ctx       context.Context
	lock      sync.Mutex

	infoHandler func(*models.InfoResponse)
	subHandler  func(*models.SubscribeResponse)

	heartbeatChan  chan *models.HeartBeat
	heartbeatTimer *time.Timer

	instrumentChan []chan<- *models.InstrumentResponse
	tradeChan      []chan<- *models.TradeResponse
	mblChan        []chan<- *models.MBLResponse

	closeFlag chan struct{}
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
func (c *Client) Connect(ctx context.Context, query string) error {
	remote := c.cfg.GetURL()

	var queryList []string

	if !strings.Contains(query, "subscribe") {
		queryList = append(queryList, "subscribe=instrument:XBTUSD,orderBookL2:XBTUSD,trade:XBTUSD")
	}
	if query != "" {
		queryList = append(queryList, query)
	}

	remote.RawQuery = strings.Join(queryList, "&")

	log.Println("Connecting to:", remote.String())

	c.ctx = ctx
	conn, rsp, err := websocket.DefaultDialer.DialContext(
		ctx, remote.String(), nil)

	if err != nil {
		return fmt.Errorf("Fail to connect[%s]: %v, %v",
			remote.String(), err, rsp)
	}

	c.ws = conn
	c.connected = true
	c.ws.SetCloseHandler(c.closeHandler)

	go c.messageHandler()
	go c.heartbeatHandler()

	return nil
}

// Closed websocket closed notification
func (c *Client) Closed() <-chan struct{} {
	return c.closeFlag
}

func (c *Client) closeHandler(code int, msg string) error {
	close(c.closeFlag)

	log.Printf("Websocket closed with code[%d]: %s\n", code, msg)

	return nil
}

// SetInfoHandler set info response handler, must be setted before calling Connect
func (c *Client) SetInfoHandler(fn func(*models.InfoResponse)) {
	if fn != nil {
		c.infoHandler = fn
	}
}

// SetSubHandler set subscribe response handler, must be setted before calling Connect & Subscribe
func (c *Client) SetSubHandler(fn func(*models.SubscribeResponse)) {
	if fn != nil {
		c.subHandler = fn
	}
}

// GetTrade to get trade data channel
func (c *Client) GetTrade() <-chan *models.TradeResponse {
	ch := make(chan *models.TradeResponse)

	c.lock.Lock()
	c.tradeChan = append(c.tradeChan, ch)
	c.lock.Unlock()

	return ch
}

// GetMBL to get mbl data channel
func (c *Client) GetMBL() <-chan *models.MBLResponse {
	ch := make(chan *models.MBLResponse)

	c.lock.Lock()
	c.mblChan = append(c.mblChan, ch)
	c.lock.Unlock()

	return ch
}

// GetInstrument to get mbl data channel
func (c *Client) GetInstrument() <-chan *models.InstrumentResponse {
	ch := make(chan *models.InstrumentResponse)

	c.lock.Lock()
	c.instrumentChan = append(c.instrumentChan, ch)
	c.lock.Unlock()

	return ch
}

func (c *Client) heartbeatHandler() {
	var heartbeatCounter int
	var err error

	if !c.cfg.ReversHeartbeat {
		go func() {
			for {
				select {
				case <-c.ctx.Done():
					return
				case <-time.NewTicker(time.Second * time.Duration(c.cfg.HeartbeatInterval)).C:
					c.heartbeatChan <- models.NewPing()
				}
			}
		}()
	} else {
		c.heartbeatTimer = time.NewTimer(time.Second * time.Duration(c.cfg.HeartbeatInterval*c.cfg.HeartbeatFailCount))

		go func() {
			for {
				select {
				case <-c.ctx.Done():
					return
				case <-c.heartbeatTimer.C:
					c.closeHandler(-1, "Receive data timeout.")

					return
				}
			}
		}()
	}

	for hb := range c.heartbeatChan {
		switch hb.Type() {
		case "Ping":
			if !c.cfg.ReversHeartbeat {
				err = c.ws.WriteMessage(websocket.TextMessage, []byte("ping"))
				heartbeatCounter += hb.Value()

				if logLevel > 0 {
					log.Println("->", hb.ToString())
				}
			} else {
				err = c.ws.WriteMessage(websocket.TextMessage, []byte("pong"))

				if logLevel > 0 {
					log.Println("<-", hb.ToString())
					log.Println("->", models.NewPong().ToString())
				}
			}
		case "Pong":
			heartbeatCounter += hb.Value()

			if logLevel > 0 {
				log.Println("<-", hb.ToString())
			}
		default:
			log.Println("Invalid heartbeat type: ", hb.ToString())

			continue
		}

		if err != nil {
			c.closeHandler(-1, "Send heartbeat failed: "+hb.ToString())

			return
		}

		if heartbeatCounter > c.cfg.HeartbeatFailCount || heartbeatCounter < 0 {
			c.closeHandler(-1, fmt.Sprint("Heartbeat miss-match:", heartbeatCounter))

			return
		}
	}
}

func (c *Client) messageHandler() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			_, msg, err := c.ws.ReadMessage()

			if err != nil {
				c.closeHandler(-1, err.Error())

				return
			}

			if c.heartbeatTimer != nil {
				c.heartbeatTimer.Reset(time.Second * time.Duration(c.cfg.HeartbeatInterval*c.cfg.HeartbeatFailCount))
			}

			if bytes.Contains(msg, pongPattern) || bytes.Contains(msg, pingPattern) {
				hbStr := string(msg)

				var hb *models.HeartBeat

				switch {
				case strings.Contains(hbStr, "ping"):
					hb = models.NewPing()
				case strings.Contains(hbStr, "pong"):
					hb = models.NewPong()
				}

				c.heartbeatChan <- hb

				continue
			}

			if bytes.Contains(msg, infoPattern) {
				var info models.InfoResponse

				if err = json.Unmarshal(msg, &info); err != nil {
					log.Println("Fail to parse info msg:", err, string(msg))

					continue
				}

				msg, _ = json.Marshal(info)

				if c.infoHandler != nil {
					c.infoHandler(&info)
				} else {
					log.Println("Info:", string(msg))
				}

				continue
			}

			if bytes.Contains(msg, subPattern) {
				var sub models.SubscribeResponse

				if err = json.Unmarshal(msg, &sub); err != nil {
					log.Println(
						"Fail to parse subscribe response:", err, string(msg))

					continue
				}

				msg, _ = json.Marshal(sub)

				if c.subHandler != nil {
					c.subHandler(&sub)
				} else {
					log.Println("Subscribe:", string(msg))
				}

				continue
			}

			switch {
			case bytes.Contains(msg, instrumentPattern):
				var insRsp models.InstrumentResponse

				if err = json.Unmarshal(msg, &insRsp); err != nil {
					log.Println(
						"Fail to parse instrument response:", err, string(msg))
					continue
				}

				msg, _ = json.Marshal(insRsp)

				for _, ch := range c.instrumentChan {
					ch <- &insRsp
				}
			case bytes.Contains(msg, tradePattern):
				var tdRsp models.TradeResponse

				if err = json.Unmarshal(msg, &tdRsp); err != nil {
					log.Println("Fail to parse trade response:", err, string(msg))
					continue
				}

				msg, _ = json.Marshal(tdRsp)

				for _, ch := range c.tradeChan {
					ch <- &tdRsp
				}
			case bytes.Contains(msg, mblPattern):
				var mblRsp models.MBLResponse

				if err = json.Unmarshal(msg, &mblRsp); err != nil {
					log.Println("Fail to parse MBL response:", err, string(msg))
					continue
				}

				msg, _ = json.Marshal(mblRsp)

				for _, ch := range c.mblChan {
					ch <- &mblRsp
				}
			default:
				log.Println("Unkonw table type:", string(msg))
			}

			if logLevel > 1 {
				log.Println("<-", string(msg))
			}
		}
	}
}

// NewClient create a new mock client instance
func NewClient(cfg *WsConfig) *Client {
	ins := Client{
		cfg:           cfg,
		heartbeatChan: make(chan *models.HeartBeat),
		closeFlag:     make(chan struct{}),
	}

	return &ins
}
