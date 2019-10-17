package server

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/frozenpine/ngerest"
	"github.com/frozenpine/wstester/models"
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
	// Welcome send welcome message to client
	Welcome() error
	// GetID to get session's unique id
	GetID() string
	// Close to close current session
	Close(int, string) error
	// Authorize to authorize current session as logged in
	Authorize(string, string)
	// IsAuthorized to specify wether current session is authrozied
	IsAuthorized() bool

	// ReadMessage receive message from client session
	ReadMessage() ([]byte, error)
	// WriteTextMessage send text message to client
	WriteTextMessage(string, bool) error
	// WriteJSONMessage send json object to client
	WriteJSONMessage(interface{}, bool) error

	// ReloadCfg reload server configurations
	ReloadCfg()
}

type message struct {
	json    interface{}
	txt     string
	errChan chan error
}

type clientSession struct {
	svr       *server
	sessionID uuid.UUID
	clientID  string
	accountID string

	conn      *websocket.Conn
	sendChan  chan *message
	isClosed  bool
	closeOnce sync.Once
	ctx       context.Context
	cancelFn  context.CancelFunc

	hbChan         chan *models.HeartBeat
	heartbeatTimer *time.Timer
}

func (s *clientSession) IsClosed() bool {
	return s.isClosed
}

func (s *clientSession) Welcome() error {
	cfg := s.getSvrCfg()
	info := models.InfoResponse{
		Info:      cfg.WelcomMsg,
		Version:   version,
		Timestamp: ngerest.NGETime(time.Now().UTC()),
		Docs:      cfg.DocsURI,
		FrontID:   cfg.FrontID,
		SessionID: s.GetID(),
	}

	return s.WriteJSONMessage(&info, true)
}

func (s *clientSession) GetID() string {
	return s.sessionID.String()
}

func (s *clientSession) Close(code int, reason string) error {
	if s.conn == nil {
		return fmt.Errorf("missing connection in current session: %s", s.GetID())
	}

	s.closeOnce.Do(func() {
		s.isClosed = true
		s.cancelFn()
		s.conn.Close()
	})

	log.Printf("Client session[%s] closed with code[%d]: %s\n", s.GetID(), code, reason)

	return nil
}

func (s *clientSession) Authorize(clientID, accountID string) {
	s.accountID = accountID
	s.clientID = clientID
}

func (s *clientSession) IsAuthorized() bool {
	return s.clientID != "" && s.accountID != ""
}

func (s *clientSession) getSvrCfg() *SvrConfig {
	return s.svr.cfg
}

func (s *clientSession) getSvcCtx() context.Context {
	return s.svr.ctx
}

func (s *clientSession) heartbeatLoop() {
	var (
		hbCounter int
		err       error
		cfg       *SvrConfig = s.getSvrCfg()
	)

	if cfg.ReversHeartbeat {
		go func() {
			ticker := time.NewTicker(time.Second * time.Duration(cfg.HeartbeatInterval))

			for {
				select {
				case <-s.ctx.Done():
					ticker.Stop()
					return
				case <-ticker.C:
					s.hbChan <- models.NewPing()
				}
			}
		}()
	} else {
		s.heartbeatTimer = time.NewTimer(time.Second * time.Duration(cfg.HeartbeatInterval*cfg.HeartbeatFailCount))

		go func() {
			for {
				select {
				case <-s.ctx.Done():
					s.heartbeatTimer.Stop()
					return
				case <-s.heartbeatTimer.C:
					s.Close(-1, "Receive data timeout.")
				}
			}
		}()
	}

	for hb := range s.hbChan {
		switch hb.Type() {
		case "Ping":
			if cfg.ReversHeartbeat {
				if err = s.WriteTextMessage("ping", true); err != nil {
					s.Close(-1, fmt.Sprintf("Send heatbeat to client session[%s] failed.", s.GetID()))
					return
				}
				hbCounter += hb.Value()
			} else {
				if err = s.WriteTextMessage("pong", true); err != nil {
					s.Close(-1, fmt.Sprintf("Send heatbeat to client session[%s] failed.", s.GetID()))
					return
				}
			}
		case "Pong":
			hbCounter += hb.Value()
		default:
			log.Println("Invalid heartbeat type: ", hb.String())

			continue
		}

		if hbCounter >= cfg.HeartbeatFailCount || hbCounter < 0 {
			s.Close(-1, fmt.Sprint("Heartbeat miss-match:", hbCounter))

			return
		}
	}
}

func (s *clientSession) ReadMessage() ([]byte, error) {
	var (
		msg []byte
		err error
		cfg *SvrConfig = s.getSvrCfg()
	)

HEARTBEAT:
	for {
		_, msg, err = s.conn.ReadMessage()

		if err != nil {
			return nil, err
		}

		if s.heartbeatTimer != nil {
			s.heartbeatTimer.Reset(time.Second * time.Duration(cfg.HeartbeatInterval*cfg.HeartbeatFailCount))
		}

		switch {
		case bytes.Contains(msg, pingPattern):
			s.hbChan <- models.NewPing()
		case bytes.Contains(msg, pongPattern):
			s.hbChan <- models.NewPong()
		default:
			break HEARTBEAT
		}
	}

	return msg, err
}

func (s *clientSession) WriteTextMessage(txt string, sync bool) (err error) {
	msg := message{txt: txt}

	if sync {
		msg.errChan = make(chan error, 1)

		defer func() {
			err = <-msg.errChan
		}()
	}

	s.sendChan <- &msg

	return
}

func (s *clientSession) WriteJSONMessage(obj interface{}, sync bool) (err error) {
	msg := message{json: obj}

	if sync {
		msg.errChan = make(chan error, 1)

		defer func() {
			err = <-msg.errChan
		}()
	}

	s.sendChan <- &msg

	return
}

func (s *clientSession) ReloadCfg() {
	// TODO: session's configuration reload
}

func (s *clientSession) sendMessageLoop() {
	var err error

	for {
		select {
		case <-s.ctx.Done():
			return
		case msg := <-s.sendChan:
			if msg == nil {
				return
			}

			if msg.json != nil {
				err = s.conn.WriteJSON(msg.json)
			} else if msg.txt != "" {
				err = s.conn.WriteMessage(websocket.TextMessage, []byte(msg.txt))
			}

			if msg.errChan != nil {
				msg.errChan <- err
			}

			if err != nil {
				s.Close(-1, err.Error())
			}
		}
	}
}

// NewSession create client session from webosocket conn
func NewSession(ctx context.Context, conn *websocket.Conn, svr *server) Session {
	sessionID := uuid.NewV3(svr.cfg.GetNS(), conn.UnderlyingConn().RemoteAddr().String())

	if ctx == nil {
		ctx = context.Background()
	}

	session := clientSession{
		svr:       svr,
		conn:      conn,
		sessionID: sessionID,
		hbChan:    make(chan *models.HeartBeat),
		sendChan:  make(chan *message),
	}

	session.ctx, session.cancelFn = context.WithCancel(ctx)

	go session.heartbeatLoop()
	go session.sendMessageLoop()

	return &session
}
