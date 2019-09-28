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

func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {

}
