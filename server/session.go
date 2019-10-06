package server

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

// Session interface interactive with client session
type Session interface {
	// GetID to get session's unique id
	GetID() string
	// Close to close current session
	Close() error
	// Authorize to authorize current session as logged in
	Authorize(string, string)
	// IsAuthorized to specify wether current session is authrozied
	IsAuthorized() bool

	WriteTextMessage(string) error
	WriteJSONMessage(interface{}) error
}

type clientSession struct {
	sessionID uuid.UUID
	clientID  string
	accountID string

	conn      *websocket.Conn
	closeOnce sync.Once
}

func (s *clientSession) GetID() string {
	return s.sessionID.String()
}

func (s *clientSession) Close() error {
	if s.conn == nil {
		return fmt.Errorf("session")
	}

	s.closeOnce.Do(func() {
		s.conn.Close()
	})

	return nil
}

func (s *clientSession) Authorize(clientID, accountID string) {
	s.accountID = accountID
	s.clientID = clientID
}

func (s *clientSession) IsAuthorized() bool {
	return s.clientID != "" && s.accountID != ""
}

func (s *clientSession) WriteTextMessage(msg string) error {
	return s.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func (s *clientSession) WriteJSONMessage(obj interface{}) error {
	return s.conn.WriteJSON(obj)
}

// NewSession create client session from webosocket conn
func NewSession(conn *websocket.Conn, ns uuid.UUID) Session {
	sessionID := uuid.NewV3(ns, conn.UnderlyingConn().RemoteAddr().String())

	session := clientSession{
		conn:      conn,
		sessionID: sessionID,
	}

	return &session
}
