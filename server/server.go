package server

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/frozenpine/wstester/models"
	"github.com/gorilla/websocket"
)

var (
	version  = "wstester mock server v0.1"
	logLevel int

	opPattern = []byte(`"op"`)
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
	Reload(*WsConfig)
}

type server struct {
	cfg      *WsConfig
	ctx      context.Context
	upgrader *websocket.Upgrader

	statics serverStatics

	clients       map[string]Session
	channelMapper map[string]Channel
}

func (s *server) Reload(cfg *WsConfig) {
	s.cfg = cfg

}

// RunForever startup and serve forever
func (s *server) RunForever(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	s.ctx = ctx

	http.HandleFunc("/status", s.statusHandler)
	http.HandleFunc(s.cfg.BaseURI, s.wsUpgrader)

	s.statics.Startup = time.Now().UTC()
	err := http.ListenAndServe(s.cfg.GetListenAddr(), nil)

	return err
}

func (s *server) incClients(conn *websocket.Conn) Session {
	session := NewSession(conn, s)
	if err := session.Welcome(); err != nil {
		log.Println(err)

		return nil
	}

	s.clients[session.GetID()] = session
	atomic.AddInt64(&s.statics.Clients, 1)

	return session
}

func (s *server) decClients(session interface{}) {
	if session == nil {
		return
	}

	var client Session

	switch session.(type) {
	case string:
		client = s.clients[session.(string)]
	case Session:
		client = session.(Session)
	}

	if client == nil {
		return
	}

	delete(s.clients, client.GetID())
	atomic.AddInt64(&s.statics.Clients, -1)
}

func (s *server) subscribe(topics ...string) {

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

func (s *server) wsUpgrader(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, w.Header())

	if err != nil {
		http.Error(w, err.Error(), 400)
	}

	clientSenssion := s.incClients(conn)
	defer func() {
		s.decClients(clientSenssion)
	}()

	var (
		msg []byte
		req models.Request
	)

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			if msg, err = clientSenssion.ReadMessage(); err != nil {
				clientSenssion.Close(-1, err.Error())
				return
			}

			switch {
			case bytes.Contains(msg, opPattern):
			default:
				log.Println("Unknow request:", string(msg))
				continue
			}
		}

		if logLevel >= 2 {
			log.Println("<-", req.String())
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
		clients: make(map[string]Session),
	}

	return &svr
}
