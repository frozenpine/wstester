package server

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/frozenpine/ngerest"
	"github.com/frozenpine/wstester/models"
	"github.com/frozenpine/wstester/utils/log"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

var (
	pingPattern = []byte(`ping`)
	pongPattern = []byte(`pong`)
)

// Session interface interactive with client session
type Session interface {
	// IsClosed session is closed.
	IsClosed() bool
	// GetAddr get client side addr.
	GetAddr() net.Addr
	// Welcome send welcome message to client
	Welcome() error
	// GetID to get session's unique id
	GetID() string
	// Close to close current session
	Close(code int, msg string) error
	// Authorize to authorize current session as logged in
	Authorize(key, secret string)
	// IsAuthorized to specify wether current session is authrozied
	IsAuthorized() bool

	// ReadMessage receive message from client session
	ReadMessage() ([]byte, error)
	// WriteTextMessage send text message to client
	WriteTextMessage(msg string, isSync bool) error
	// WriteJSONMessage send json object to client
	WriteJSONMessage(obj interface{}, isSync bool) error
}

type message struct {
	json    interface{}
	txt     string
	errChan chan error
}

type clientSession struct {
	cfg       *Config
	sessionID uuid.UUID
	clientID  string
	accountID string

	conn      *websocket.Conn
	req       *http.Request
	addr      net.Addr
	sendChan  chan *message
	isClosed  bool
	closeOnce sync.Once
	ctx       context.Context
	cancelFn  context.CancelFunc

	hbChan         chan *models.HeartBeat
	heartbeatTimer *time.Timer
}

func (c *clientSession) IsClosed() bool {
	return c.isClosed
}

func (c *clientSession) Welcome() error {
	info := models.InfoResponse{
		Info:      c.cfg.WelcomMsg,
		Version:   version,
		Timestamp: ngerest.NGETime(time.Now().UTC()),
		Docs:      c.cfg.DocsURI,
		FrontID:   c.cfg.FrontID,
		SessionID: c.GetID(),
	}

	return c.WriteJSONMessage(&info, true)
}

func (c *clientSession) GetID() string {
	return c.sessionID.String()
}

func (c *clientSession) GetAddr() net.Addr {
	return c.addr
}

func (c *clientSession) Close(code int, reason string) error {
	if c.conn == nil {
		return fmt.Errorf("missing connection in current session: %s", c.GetID())
	}

	c.closeOnce.Do(func() {
		c.isClosed = true
		c.cancelFn()
		c.conn.Close()
	})

	log.Infof("Client session[%s] closed with code[%d]: %s", c.GetID(), code, reason)

	return nil
}

func (c *clientSession) Authorize(clientID, accountID string) {
	c.accountID = accountID
	c.clientID = clientID
}

func (c *clientSession) IsAuthorized() bool {
	return c.clientID != "" && c.accountID != ""
}

func (c *clientSession) heartbeatLoop() {
	var (
		hbCounter int
		err       error
	)

	if c.cfg.ReversHeartbeat {
		go func() {
			ticker := time.NewTicker(time.Second * time.Duration(c.cfg.HeartbeatInterval))

			for {
				select {
				case <-c.ctx.Done():
					ticker.Stop()
					return
				case <-ticker.C:
					c.hbChan <- models.NewPing()
				}
			}
		}()
	} else {
		c.heartbeatTimer = time.NewTimer(time.Second * time.Duration(c.cfg.HeartbeatInterval*c.cfg.HeartbeatFailCount))

		go func() {
			for {
				select {
				case <-c.ctx.Done():
					c.heartbeatTimer.Stop()
					return
				case <-c.heartbeatTimer.C:
					c.Close(-1, "Receive data timeout.")
				}
			}
		}()
	}

	for hb := range c.hbChan {
		switch hb.Type() {
		case "Ping":
			if c.cfg.ReversHeartbeat {
				if err = c.WriteTextMessage("ping", true); err != nil {
					c.Close(-1, fmt.Sprintf("Send heatbeat to client session[%s] failed.", c.GetID()))
					return
				}
				hbCounter += hb.Value()
			} else {
				if err = c.WriteTextMessage("pong", true); err != nil {
					c.Close(-1, fmt.Sprintf("Send heatbeat to client session[%s] failed.", c.GetID()))
					return
				}
			}
		case "Pong":
			hbCounter += hb.Value()
		default:
			log.Error("Invalid heartbeat type: ", hb.String())

			continue
		}

		if hbCounter >= c.cfg.HeartbeatFailCount || hbCounter < 0 {
			c.Close(-1, fmt.Sprint("Heartbeat miss-match:", hbCounter))

			return
		}
	}
}

func (c *clientSession) ReadMessage() ([]byte, error) {
	var (
		msg []byte
		err error
	)

HEARTBEAT:
	for {
		_, msg, err = c.conn.ReadMessage()

		if err != nil {
			return nil, err
		}

		if c.heartbeatTimer != nil {
			c.heartbeatTimer.Reset(time.Second * time.Duration(c.cfg.HeartbeatInterval*c.cfg.HeartbeatFailCount))
		}

		switch {
		case bytes.Contains(msg, pingPattern):
			c.hbChan <- models.NewPing()
		case bytes.Contains(msg, pongPattern):
			c.hbChan <- models.NewPong()
		default:
			break HEARTBEAT
		}
	}

	return msg, err
}

func (c *clientSession) WriteTextMessage(txt string, sync bool) (err error) {
	msg := message{txt: txt}

	if sync {
		msg.errChan = make(chan error, 1)

		defer func() {
			err = <-msg.errChan
		}()
	}

	c.sendChan <- &msg

	return
}

func (c *clientSession) WriteJSONMessage(obj interface{}, sync bool) (err error) {
	msg := message{json: obj}

	if sync {
		msg.errChan = make(chan error, 1)

		defer func() {
			err = <-msg.errChan
		}()
	}

	c.sendChan <- &msg

	return
}

func (c *clientSession) sendMessageLoop() {
	var err error

	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.sendChan:
			if msg == nil {
				return
			}

			if msg.json != nil {
				err = c.conn.WriteJSON(msg.json)
			} else if msg.txt != "" {
				err = c.conn.WriteMessage(websocket.TextMessage, []byte(msg.txt))
			}

			if msg.errChan != nil {
				msg.errChan <- err
			}

			if err != nil {
				c.Close(-1, err.Error())
			}
		}
	}
}

// NewSession create client session from webosocket conn
func NewSession(ctx context.Context, conn *websocket.Conn, req *http.Request) Session {
	cfg := ctx.Value(SvrConfigKey).(*Config)
	// TODO: get real addr x-forwared-for from upgrade request.
	sessionID := uuid.NewV3(cfg.GetNS(), conn.RemoteAddr().String())

	if ctx == nil {
		ctx = context.Background()
	}

	session := clientSession{
		cfg:       cfg,
		req:       req,
		conn:      conn,
		addr:      conn.RemoteAddr(),
		sessionID: sessionID,
		hbChan:    make(chan *models.HeartBeat),
		sendChan:  make(chan *message),
	}

	session.ctx, session.cancelFn = context.WithCancel(ctx)

	go session.heartbeatLoop()
	go session.sendMessageLoop()

	return &session
}
