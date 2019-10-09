package models

import (
	"encoding/json"
	"regexp"

	"github.com/frozenpine/ngerest"
)

var (
	// PingPattern ping message pattern
	PingPattern = regexp.MustCompile(`ping`)

	// PongPattern pong message pattern
	PongPattern = regexp.MustCompile(`pong`)

	// InfoPattern info message pattern
	InfoPattern = regexp.MustCompile(`"info"`)

	// SubPattern subscribe message pattern
	SubPattern = regexp.MustCompile(`"subscribe"`)

	// AuthPattern auth message pattern
	AuthPattern = regexp.MustCompile(`"authKeyExpires"|"api-key"`)

	// InstrumentPattern instrument message pattern
	InstrumentPattern = regexp.MustCompile(`"table": ?"instrument"`)

	// MBLPattern mbl message pattern
	MBLPattern = regexp.MustCompile(`"table": ?"orderBook`)

	// TradePattern trade message pattern
	TradePattern = regexp.MustCompile(`"table": ?"trade"`)

	// ErrPattern error message pattern
	ErrPattern = regexp.MustCompile(`"error"`)
)

// Response common functions for response
type Response interface {
	String() string
	Format(string) string
	IsTableResponse() bool
	IsPartialResponse() bool
}

// TableResponse table data response
type TableResponse interface {
	Response

	GetAction() string
	GetData() []interface{}
}

// Request common functions for request
type Request interface {
	String() string
	GetOperation() string
	GetArgs() []string
}

// OperationRequest request to websocket
type OperationRequest struct {
	Operation string   `json:"op"`
	Args      []string `json:"args"`
}

func (req *OperationRequest) String() string {
	result, _ := json.Marshal(req)

	return string(result)
}

// GetOperation get request's operation name
func (req *OperationRequest) GetOperation() string {
	return req.Operation
}

// GetArgs get request's args slice
func (req *OperationRequest) GetArgs() []string {
	return req.Args
}

// InfoResponse welcome message
type InfoResponse struct {
	Info      string                 `json:"info"`
	Version   string                 `json:"version"`
	Timestamp ngerest.NGETime        `json:"timestamp"`
	Docs      string                 `json:"docs"`
	Limit     map[string]interface{} `json:"limit"`
	FrontID   string                 `json:"frontId"`
	SessionID string                 `json:"sessionId"`
}

// String get structure's string format
func (info *InfoResponse) String() string {
	result, _ := json.Marshal(info)

	return string(result)
}

// Format format String output
func (info *InfoResponse) Format(format string) string {
	return info.String()
}

// IsTableResponse determinate wether response is a table data
func (info *InfoResponse) IsTableResponse() bool {
	return false
}

// IsPartialResponse determinate wether table response is partial data
func (info *InfoResponse) IsPartialResponse() bool {
	return false
}

// AuthResponse authentication response
type AuthResponse struct {
	Success bool                   `json:"success"`
	Request map[string]interface{} `json:"request"`
}

// String get structure's string format
func (auth *AuthResponse) String() string {
	result, _ := json.Marshal(auth)

	return string(result)
}

// Format format String output
func (auth *AuthResponse) Format(format string) string {
	return auth.String()
}

// IsTableResponse determinate wether response is a table data
func (auth *AuthResponse) IsTableResponse() bool {
	return false
}

// IsPartialResponse determinate wether table response is partial data
func (auth *AuthResponse) IsPartialResponse() bool {
	return false
}

// SubscribeResponse subscribe response
type SubscribeResponse struct {
	Success   bool             `json:"success"`
	Subscribe string           `json:"subscribe"`
	Request   OperationRequest `json:"request"`
}

// String get structure's string format
func (sub *SubscribeResponse) String() string {
	result, _ := json.Marshal(sub)

	return string(result)
}

// Format format String output
func (sub *SubscribeResponse) Format(format string) string {
	return sub.String()
}

// IsTableResponse determinate wether response is a table data
func (sub *SubscribeResponse) IsTableResponse() bool {
	return false
}

// IsPartialResponse determinate wether table response is partial data
func (sub *SubscribeResponse) IsPartialResponse() bool {
	return false
}

// ErrResponse error response
type ErrResponse struct {
	Error string `json:"error"`

	Status  int                    `json:"status,omitempty"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
	Request OperationRequest       `json:"request,omitempty"`
}

// String get structure's string format
func (err *ErrResponse) String() string {
	result, _ := json.Marshal(err)

	return string(result)
}

// Format format String output
func (err *ErrResponse) Format(format string) string {
	return err.String()
}

// IsTableResponse determinate wether response is a table data
func (err *ErrResponse) IsTableResponse() bool {
	return false
}

// IsPartialResponse determinate wether table response is partial data
func (err *ErrResponse) IsPartialResponse() bool {
	return false
}

type tableResponse struct {
	Table  string `json:"table"`
	Action string `json:"action"`

	Keys        []string          `json:"keys,omitempty"`
	Types       map[string]string `json:"types,omitempty"`
	ForeignKeys map[string]string `json:"foreignKeys,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	Filter      map[string]string `json:"filter,omitempty"`
}

// IsTableResponse determinate wether response is a table data
func (tbl *tableResponse) IsTableResponse() bool {
	return true
}

// IsPartialResponse determinate wether table response is partial data
func (tbl *tableResponse) IsPartialResponse() bool {
	return tbl.Action == "partial"
}
