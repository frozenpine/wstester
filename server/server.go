package server

import (
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
		Info:      "Welcom to NGE websocket service mock.",
		Version:   "Mock Server v1",
		Timestamp: ngerest.NGETime(time.Now().UTC()),
		Docs:      "https://docs.btcmex.com",
		FrontID:   "0",
		SessionID: "123456",
	}

	conn.WriteJSON(info)

	// var (
	// 	msg []byte
	// )

	// for {
	// 	_, msg, err = conn.ReadMessage()

	// 	if err != nil {
	// 		return
	// 	}
	// }
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
