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
	WriteTextMessage(string) error
	// WriteJSONMessage send json object to client
	WriteJSONMessage(interface{}) error

	// ReloadCfg reload server configurations
	ReloadCfg()
}

type clientSession struct {
	svr       *server
	sessionID uuid.UUID
	clientID  string
	accountID string

	conn      *websocket.Conn
	closeOnce sync.Once

	hbChan         chan *models.HeartBeat
	heartbeatTimer *time.Timer
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

	return s.WriteJSONMessage(&info)
}

func (s *clientSession) GetID() string {
	return s.sessionID.String()
}

func (s *clientSession) Close(code int, reason string) error {
	if s.conn == nil {
		return fmt.Errorf("missing connection in current session: %s", s.GetID())
	}

	s.closeOnce.Do(func() {
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

func (s *clientSession) getSvrCfg() *WsConfig {
	return s.svr.cfg
}

func (s *clientSession) getSvcCtx() context.Context {
	return s.svr.ctx
}

func (s *clientSession) heartbeatHandler() {
	var (
		hbCounter int
		err       error
		cfg       *WsConfig       = s.getSvrCfg()
		ctx       context.Context = s.getSvcCtx()
	)

	if cfg.ReversHeartbeat {
		go func() {
			for {
				select {
				case <-ctx.Done():
					s.Close(0, "Server exit.")
					return
				case <-time.NewTicker(time.Second * time.Duration(cfg.HeartbeatInterval)).C:
					s.hbChan <- models.NewPing()
				}
			}
		}()
	} else {
		s.heartbeatTimer = time.NewTimer(time.Second * time.Duration(cfg.HeartbeatInterval*cfg.HeartbeatFailCount))

		go func() {
			for {
				select {
				case <-ctx.Done():
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
				if err = s.WriteTextMessage("ping"); err != nil {
					s.Close(-1, fmt.Sprintf("Send heatbeat to client session[%s] failed.", s.GetID()))
					return
				}
				hbCounter += hb.Value()
			} else {
				if err = s.WriteTextMessage("pong"); err != nil {
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
		cfg *WsConfig = s.getSvrCfg()
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

func (s *clientSession) WriteTextMessage(msg string) error {
	return s.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func (s *clientSession) WriteJSONMessage(obj interface{}) error {
	return s.conn.WriteJSON(obj)
}

func (s *clientSession) ReloadCfg() {
	// TODO: session's configuration reload
}

// NewSession create client session from webosocket conn
func NewSession(conn *websocket.Conn, svr *server) Session {
	sessionID := uuid.NewV3(svr.cfg.GetNS(), conn.UnderlyingConn().RemoteAddr().String())

	session := clientSession{
		svr:       svr,
		conn:      conn,
		sessionID: sessionID,
		hbChan:    make(chan *models.HeartBeat),
	}

	go session.heartbeatHandler()

	return &session
}
