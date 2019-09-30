package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:    4096,
	WriteBufferSize:   4096,
	EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Status statics for running server
type Status struct {
	Clients int `json:"clients"`
}

// Server server instance
type Server interface {
	RunForever() error
}

type server struct {
	cfg      *WsConfig
	upgrader *websocket.Upgrader

	status Status
}

// RunForever startup and serve forever
func (s *server) RunForever() error {
	http.HandleFunc("/status", s.statusHandler)
	http.HandleFunc(s.cfg.BaseURI, s.wsHandler)

	err := http.ListenAndServe(s.cfg.GetListenAddr(), nil)

	return err
}

func (s *server) wsHandler(w http.ResponseWriter, r *http.Request) {
	_, err := upgrader.Upgrade(w, r, w.Header())

	if err != nil {
		http.Error(w, err.Error(), 400)
	}
}

func (s *server) statusHandler(w http.ResponseWriter, r *http.Request) {
	status, _ := json.Marshal(s.status)

	w.Header().Set("Content-type", "application/json")
	w.Write(status)
}
