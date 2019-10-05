package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/frozenpine/ngerest"
	"github.com/frozenpine/wstester/models"
	"github.com/gorilla/websocket"
)

var (
	pingPattern = []byte(`ping`)
	pongPattern = []byte(`pong`)
	opPattern   = []byte(`"op"`)
)

type serverStatics struct {
	Startup time.Time `json:"startup"`
	Clients int64     `json:"clients"`
}

// Status statics for running server
type Status struct {
	serverStatics

	Uptime string `json:"uptime"`
}

// Server server instance
type Server interface {
	RunForever(ctx context.Context) error
}

type server struct {
	cfg      *WsConfig
	ctx      context.Context
	upgrader *websocket.Upgrader

	statics serverStatics
}

// RunForever startup and serve forever
func (s *server) RunForever(ctx context.Context) error {
	http.HandleFunc("/status", s.statusHandler)
	http.HandleFunc(s.cfg.BaseURI, s.wsUpgrader)

	s.statics.Startup = time.Now().UTC()
	err := http.ListenAndServe(s.cfg.GetListenAddr(), nil)

	return err
}

func (s *server) incClients(conn *websocket.Conn) {
	// remoteAddr := conn.UnderlyingConn().RemoteAddr().String()

	atomic.AddInt64(&s.statics.Clients, 1)
}

func (s *server) decClients(conn *websocket.Conn) {
	atomic.AddInt64(&s.statics.Clients, -1)
}

func (s *server) checkAuthHeader(r *http.Request) error {
	apiKey := r.Header.Get("api-key")
	if apiKey == "" {

	}

	apiSignature := r.Header.Get("api-signature")
	if apiSignature == "" {

	}

	apiExpires, err := strconv.ParseInt(r.Header.Get("api-expires"), 10, 64)
	if err != nil {

	}
	if apiExpires > time.Now().Unix() {
		return NewAPIExpires(apiExpires)
	}

	return nil
}

func (s *server) WriteTextMessage(conn *websocket.Conn, msg string) error {
	return conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func (s *server) heartbeatHandler(conn *websocket.Conn, hbChan chan *models.HeartBeat) {
	var hbCounter int

	for hb := range hbChan {
		switch hb.Type() {
		case "Ping":
			if s.cfg.ReversHeartbeat {
				s.WriteTextMessage(conn, "ping")
				hbCounter += hb.Value()
			} else {
				s.WriteTextMessage(conn, "pong")
			}
		case "Pong":
			hbCounter += hb.Value()
		default:
		}
	}
}

func (s *server) readMessage(conn *websocket.Conn) ([]byte, error) {
	var (
		hbChan chan *models.HeartBeat
		msg    []byte
		err    error
	)

	go s.heartbeatHandler(conn, hbChan)

HEARTBEAT:
	for {
		_, msg, err = conn.ReadMessage()

		if err != nil {
			return nil, err
		}

		switch {
		case bytes.Contains(msg, pingPattern):
			hbChan <- models.NewPing()
		case bytes.Contains(msg, pongPattern):
			hbChan <- models.NewPong()
		default:
			break HEARTBEAT
		}
	}

	return msg, err
}

func (s *server) wsUpgrader(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, w.Header())

	if err != nil {
		http.Error(w, err.Error(), 400)
	}

	s.incClients(conn)
	defer func() {
		s.decClients(conn)
	}()

	info := models.InfoResponse{
		Info:      s.cfg.WelcomMsg,
		Version:   "Mock Server v1",
		Timestamp: ngerest.NGETime(time.Now().UTC()),
		Docs:      s.cfg.DocsURI,
		FrontID:   s.cfg.FrontID,
		SessionID: "123456",
	}

	conn.WriteJSON(&info)

	var (
		msg []byte
	)

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			if msg, err = s.readMessage(conn); err != nil {
				return
			}

			switch {
			case bytes.Contains(msg, opPattern):
			default:
			}
		}
	}
}

func (s *server) statusHandler(w http.ResponseWriter, r *http.Request) {
	status := Status{
		serverStatics: s.statics,
		Uptime:        time.Now().Sub(s.statics.Startup).String(),
	}
	statusResult, _ := json.Marshal(status)

	w.Header().Set("Content-type", "application/json")
	w.Write(statusResult)
}

//NewServer to create a websocket server
func NewServer(cfg *WsConfig) Server {
	svr := server{
		cfg: cfg,
		upgrader: &websocket.Upgrader{
			ReadBufferSize:    4096,
			WriteBufferSize:   4096,
			EnableCompression: true,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		statics: serverStatics{},
	}

	return &svr
}
