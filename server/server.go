package server

import (
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

// Server server instance
type Server struct {
	cfg      *WsConfig
	upgrader *websocket.Upgrader
}

// RunForever startup and serve forever
func (s *Server) RunForever() error {
	http.HandleFunc("/status", s.statusHandler)
	http.HandleFunc(s.cfg.BaseURI, s.wsHandler)

	err := http.ListenAndServe(s.cfg.GetListenAddr(), nil)

	return err
}

func (s *Server) wsHandler(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {

}
