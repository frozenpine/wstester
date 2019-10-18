package kafka

import (
	"encoding/json"
	"errors"
	"regexp"

	"github.com/frozenpine/wstester/models"
)

var (
	marginNotifyPattern   = regexp.MustCompile(`"type": ?"margin"`)
	positionNotifyPattern = regexp.MustCompile(`"type": ?"position"`)
	execNotifyPattern     = regexp.MustCompile(`"type": ?"execution"`)
	orderNotifyPattern    = regexp.MustCompile(`"type": ?"order"`)
	tradeNotifyPattern    = regexp.MustCompile(`"type": ?"trade"`)
)

// Notify message in kafka NOTIFY topic
type Notify interface {
	// IsPrivate wether this notify message is a private flow
	IsPrivate() bool
	// GetClientID get client id in notify
	GetClientID() string
	// GetAccountID get account id in notify
	GetAccountID() string
	// GetContent get content data in notify
	GetContent() interface{}
}

// ParseNotify parse notify message
func ParseNotify(notify []byte) (msg Notify, err error) {
	switch {
	case marginNotifyPattern.Match(notify):
		margin := MarginNotify{}
		if err = json.Unmarshal(notify, &margin); err == nil {
			msg = &margin
		}
	case positionNotifyPattern.Match(notify):
		pos := PositionNotify{}
		if err = json.Unmarshal(notify, &pos); err == nil {
			msg = &pos
		}
	case execNotifyPattern.Match(notify):
		exec := ExecutionNotify{}
		if err = json.Unmarshal(notify, &exec); err == nil {
			msg = &exec
		}
	case orderNotifyPattern.Match(notify):
		ord := OrderNotify{}
		if err = json.Unmarshal(notify, &ord); err == nil {
			msg = &ord
		}
	case tradeNotifyPattern.Match(notify):
		trade := TradeNotify{}
		if err = json.Unmarshal(notify, &trade); err == nil {
			msg = &trade
		}
	default:
		err = errors.New("Unkonw notify message: " + string(notify))
	}

	return
}

type notifyMessage struct {
	Type      string `json:"type"`
	ClientID  string `json:"clientId,omitempty"`
	AccountID string `json:"accountId,omitempty"`
}

func (n *notifyMessage) IsPrivate() bool {
	return n.ClientID != "" && n.AccountID != ""
}

func (n *notifyMessage) GetClientID() string {
	return n.ClientID
}

func (n *notifyMessage) GetAccountID() string {
	return n.AccountID
}

// MarginNotify notify message for margin
type MarginNotify struct {
	notifyMessage

	Content *models.MarginResponse `json:"content"`
}

// GetContent get content in margin notify
func (margin *MarginNotify) GetContent() interface{} {
	return margin.Content
}

// PositionNotify notify message for margin
type PositionNotify struct {
	notifyMessage

	Content *models.PositionResponse `json:"content"`
}

// GetContent get content in position notify
func (pos *PositionNotify) GetContent() interface{} {
	return pos.Content
}

// ExecutionNotify notify message for margin
type ExecutionNotify struct {
	notifyMessage

	Content *models.ExecutionResponse `json:"content"`
}

// GetContent get content in execution notify
func (exec *ExecutionNotify) GetContent() interface{} {
	return exec.Content
}

// OrderNotify notify message for margin
type OrderNotify struct {
	notifyMessage

	Content *models.OrderResponse `json:"content"`
}

// GetContent get content in order notify
func (ord *OrderNotify) GetContent() interface{} {
	return ord.Content
}

// TradeNotify notify message for trade
type TradeNotify struct {
	notifyMessage

	Content *models.TradeResponse `json:"content"`
}

// GetContent get content in trade notify
func (td *TradeNotify) GetContent() interface{} {
	return td.Content
}

// MBLNotify notify message for trade
type MBLNotify struct {
	notifyMessage

	Content *models.MBLResponse `json:"content"`
}

// GetContent get content in mbl notify
func (mbl *MBLNotify) GetContent() interface{} {
	return mbl.Content
}
