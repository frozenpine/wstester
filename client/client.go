package client

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/frozenpine/wstester/models"
	"github.com/gorilla/websocket"
)

var (
	pingPattern       = []byte(`ping`)
	pongPattern       = []byte(`pong`)
	infoPattern       = []byte(`"info"`)
	instrumentPattern = []byte(`"instrument"`)
	tradePattern      = []byte(`"trade"`)
	mblPattern        = []byte(`"orderBook`)
	subPattern        = []byte(`"subscribe"`)
	authPattern1      = []byte(`"authKeyExpires"`)
	authPattern2      = []byte(`"api-key"`)
)

// Client client instance
type Client interface {
	Host() string
	Connect(ctx context.Context) error
	Closed() <-chan bool
	Subscribe(topics ...string)
	UnSubscribe(topics ...string)
	IsConnected() bool
	IsAuthencated() bool
	SendJSONMessage(msg interface{}) error
	SetInfoHandler(func(*models.InfoResponse))
	SetSubHandler(func(*models.SubscribeResponse))
	GetTrade() <-chan *models.TradeResponse
	GetMBL() <-chan *models.MBLResponse
	GetInstrument() <-chan *models.InstrumentResponse
}

type client struct {
	cfg         *WsConfig
	ws          *websocket.Conn
	connected   bool
	authencated bool
	ctx         context.Context
	lock        sync.Mutex

	infoHandler func(*models.InfoResponse)
	subHandler  func(*models.SubscribeResponse)

	heartbeatChan  chan *models.HeartBeat
	heartbeatTimer *time.Timer

	instrumentChan []chan<- *models.InstrumentResponse
	tradeChan      []chan<- *models.TradeResponse
	mblChan        []chan<- *models.MBLResponse

	closeFlag chan bool

	SubscribedTopics map[string]*models.SubscribeResponse
}

// Host to get remote host string
func (c *client) Host() string {
	return c.cfg.GetURL().String()
}

// IsConnected to specify if client is connected to remote host
func (c *client) IsConnected() bool {
	return c.connected && c.ws != nil
}

// IsAuthencated to specify if client is logged in to remote host
func (c *client) IsAuthencated() bool {
	return c.authencated && c.ws != nil
}

func (c *client) hasAuth() bool {
	if c.ctx == nil {
		return false
	}

	_, exist := c.ctx.Value(ContextAPIKey).(APIKeyAuth)

	return exist
}

func (c *client) getAuth() *APIKeyAuth {
	if c.ctx == nil {
		return nil
	}

	if auth, ok := c.ctx.Value(ContextAPIKey).(APIKeyAuth); ok {
		return &auth
	}

	return nil
}

func (c *client) getHeader() http.Header {
	if c.ctx == nil {
		return nil
	}

	if auth := c.getAuth(); auth != nil {
		header := make(http.Header)

		header["api-key"] = []string{auth.Key}

		remote := c.cfg.GetURL()
		remote.Path = auth.AuthURI

		nonce := int(time.Now().Unix() + 5)

		header["api-signature"] = []string{c.generateSignature(
			auth.Secret, "get", remote, nonce, nil)}

		header["api-expires"] = []string{strconv.Itoa(nonce)}

		return header
	}

	return nil
}

func (c *client) isSubscribed(topic string) bool {
	rsp, exist := c.SubscribedTopics[topic]

	return exist && rsp != nil && rsp.Success
}

func (c *client) normalizeTopic(topic string) string {
	for _, name := range symbolSubs {
		if strings.ToLower(name) == strings.ToLower(topic) {
			return strings.Join([]string{name, c.cfg.Symbol}, ":")
		}
	}

	for _, name := range append(PublicTopics, PrivateTopics...) {
		if strings.ToLower(name) == strings.ToLower(topic) {
			return name
		}
	}

	return ""
}

// Subscribe subscribe topic
func (c *client) Subscribe(topics ...string) {
	var subArgs []string

	defer func() {
		if c.connected {
			sub := models.OperationRequest{
				Operation: "subscribe",
				Args:      subArgs,
			}

			c.SendJSONMessage(sub)
		}
	}()

	for _, topic := range topics {
		if !IsValidTopic(topic) {
			log.Println("Invalid topic name:", topic)
			continue
		}

		if c.isSubscribed(topic) {
			log.Printf("Topic[%s] already subscirbed.\n", topic)
			continue
		}

		c.SubscribedTopics[topic] = nil

		subArgs = append(subArgs, c.normalizeTopic(topic))
	}
}

// UnSubscribe unsubscribe topic
func (c *client) UnSubscribe(topics ...string) {
	for _, topic := range topics {
		if !IsValidTopic(topic) {
			log.Println("Invalid topic name:", topic)
			continue
		}

		if !c.isSubscribed(topic) {
			log.Printf("Topic[%s] is not subscribed.\n", topic)
			continue
		}

		c.SubscribedTopics[topic] = nil

		// TODO: real unsubscribe action
	}
}

// Connect to remote host
func (c *client) Connect(ctx context.Context) error {
	remote := c.cfg.GetURL()

	var subList []string

	for topic := range c.SubscribedTopics {
		if IsPublicTopic(topic) {
			subList = append(subList, c.normalizeTopic(topic))

			continue
		}

		if IsPrivateTopic(topic) && c.hasAuth() {
			subList = append(subList, c.normalizeTopic(topic))
		}
	}

	if len(subList) > 0 {
		remote.RawQuery = "subscribe=" + strings.Join(subList, ",")
	}

	log.Println("Connecting to:", remote.String())

	c.ctx = ctx
	conn, rsp, err := websocket.DefaultDialer.DialContext(
		ctx, remote.String(), c.getHeader())

	if err != nil {
		return fmt.Errorf("Fail to connect[%s]: %v, %v",
			remote.String(), err, rsp)
	}

	defer func() {
		go c.messageHandler()
		go c.heartbeatHandler()
	}()

	c.ws = conn
	c.connected = true
	c.ws.SetCloseHandler(c.closeHandler)

	return nil
}

// Closed websocket closed notification
func (c *client) Closed() <-chan bool {
	return c.closeFlag
}

func (c *client) closeHandler(code int, msg string) error {
	close(c.closeFlag)

	c.connected = false
	c.authencated = false

	log.Printf("Websocket closed with code[%d]: %s\n", code, msg)

	return nil
}

// SendJSONMessage send json message to remote
func (c *client) SendJSONMessage(msg interface{}) error {
	return c.ws.WriteJSON(msg)
}

// SetInfoHandler set info response handler, must be setted before calling Connect
func (c *client) SetInfoHandler(fn func(*models.InfoResponse)) {
	if fn != nil {
		c.infoHandler = fn
	}
}

// SetSubHandler set subscribe response handler, must be setted before calling Connect & Subscribe
func (c *client) SetSubHandler(fn func(*models.SubscribeResponse)) {
	if fn != nil {
		c.subHandler = fn
	}
}

// GetTrade to get trade data channel
func (c *client) GetTrade() <-chan *models.TradeResponse {
	ch := make(chan *models.TradeResponse)

	c.lock.Lock()
	c.tradeChan = append(c.tradeChan, ch)
	c.lock.Unlock()

	return ch
}

// GetMBL to get mbl data channel
func (c *client) GetMBL() <-chan *models.MBLResponse {
	ch := make(chan *models.MBLResponse)

	c.lock.Lock()
	c.mblChan = append(c.mblChan, ch)
	c.lock.Unlock()

	return ch
}

// GetInstrument to get mbl data channel
func (c *client) GetInstrument() <-chan *models.InstrumentResponse {
	ch := make(chan *models.InstrumentResponse)

	c.lock.Lock()
	c.instrumentChan = append(c.instrumentChan, ch)
	c.lock.Unlock()

	return ch
}

func (c *client) heartbeatHandler() {
	var heartbeatCounter int
	var err error

	if !c.cfg.ReversHeartbeat {
		go func() {
			for {
				select {
				case <-c.ctx.Done():
					return
				case <-time.NewTicker(time.Second * time.Duration(c.cfg.HeartbeatInterval)).C:
					c.heartbeatChan <- models.NewPing()
				}
			}
		}()
	} else {
		c.heartbeatTimer = time.NewTimer(time.Second * time.Duration(c.cfg.HeartbeatInterval*c.cfg.HeartbeatFailCount))

		go func() {
			for {
				select {
				case <-c.ctx.Done():
					return
				case <-c.heartbeatTimer.C:
					c.closeHandler(-1, "Receive data timeout.")
					return
				}
			}
		}()
	}

	for hb := range c.heartbeatChan {
		switch hb.Type() {
		case "Ping":
			if !c.cfg.ReversHeartbeat {
				err = c.ws.WriteMessage(websocket.TextMessage, []byte("ping"))
				heartbeatCounter += hb.Value()

				if logLevel >= 1 {
					log.Println("->", hb.ToString())
				}
			} else {
				err = c.ws.WriteMessage(websocket.TextMessage, []byte("pong"))

				if logLevel >= 1 {
					log.Println("<-", hb.ToString())
					log.Println("->", models.NewPong().ToString())
				}
			}
		case "Pong":
			heartbeatCounter += hb.Value()

			if logLevel > 0 {
				log.Println("<-", hb.ToString())
			}
		default:
			log.Println("Invalid heartbeat type: ", hb.ToString())

			continue
		}

		if err != nil {
			c.closeHandler(-1, "Send heartbeat failed: "+hb.ToString())

			return
		}

		if heartbeatCounter > c.cfg.HeartbeatFailCount || heartbeatCounter < 0 {
			c.closeHandler(-1, fmt.Sprint("Heartbeat miss-match:", heartbeatCounter))

			return
		}
	}
}

func (c *client) readMessage() ([]byte, error) {
	var (
		msg []byte
		err error
	)

HEARTBEAT:
	for {
		_, msg, err = c.ws.ReadMessage()

		if err != nil {
			return nil, err
		}

		if c.heartbeatTimer != nil {
			c.heartbeatTimer.Reset(time.Second * time.Duration(c.cfg.HeartbeatInterval*c.cfg.HeartbeatFailCount))
		}

		switch {
		case bytes.Contains(msg, pongPattern):
			c.heartbeatChan <- models.NewPong()
		case bytes.Contains(msg, pingPattern):
			c.heartbeatChan <- models.NewPing()
		default:
			break HEARTBEAT
		}
	}

	return msg, err
}

func (c *client) handleInfoMsg(msg []byte) (*models.InfoResponse, error) {
	var info models.InfoResponse

	if err := json.Unmarshal(msg, &info); err != nil {
		return nil, err
	}

	defer func() {
		if c.infoHandler != nil {
			c.infoHandler(&info)
		} else {
			log.Println("Info:", info.ToString())
		}
	}()

	return &info, nil
}

func (c *client) handleAuthMsg(msg []byte) (*models.AuthResponse, error) {
	var auth models.AuthResponse

	if err := json.Unmarshal(msg, &auth); err != nil {
		return nil, err
	}

	defer func() {
		if auth.Success {
			c.authencated = true
		}

		log.Println("Auth:", auth.ToString())
	}()

	return &auth, nil
}

func (c *client) handlSubMsg(msg []byte) (*models.SubscribeResponse, error) {
	var sub models.SubscribeResponse

	if err := json.Unmarshal(msg, &sub); err != nil {
		return nil, err
	}

	defer func() {
		topic := strings.Split(sub.Subscribe, ":")[0]

		if sub.Success {
			c.SubscribedTopics[topic] = &sub
		} else {
			delete(c.SubscribedTopics, topic)
		}

		if c.subHandler != nil {
			c.subHandler(&sub)
		} else {
			log.Println("Subscribe:", sub.ToString())
		}
	}()

	return &sub, nil
}

func (c *client) handleInsMsg(msg []byte) (*models.InstrumentResponse, error) {
	var insRsp models.InstrumentResponse

	if err := json.Unmarshal(msg, &insRsp); err != nil {
		return nil, err
	}

	defer func() {
		if len(c.instrumentChan) < 1 {
			return
		}

		for _, ch := range c.instrumentChan {
			ch <- &insRsp
		}
	}()

	return &insRsp, nil
}

func (c *client) handleTdMsg(msg []byte) (*models.TradeResponse, error) {
	var tdRsp models.TradeResponse

	if err := json.Unmarshal(msg, &tdRsp); err != nil {
		return nil, err
	}

	defer func() {
		if len(c.tradeChan) < 1 {
			return
		}

		for _, ch := range c.tradeChan {
			ch <- &tdRsp
		}
	}()

	return &tdRsp, nil
}

func (c *client) handleMblMsg(msg []byte) (*models.MBLResponse, error) {
	var mblRsp models.MBLResponse

	if err := json.Unmarshal(msg, &mblRsp); err != nil {
		return nil, err
	}

	defer func() {
		if len(c.mblChan) < 1 {
			return
		}

		for _, ch := range c.mblChan {
			ch <- &mblRsp
		}
	}()

	return &mblRsp, nil
}

func (c *client) messageHandler() {
	var (
		msg []byte
		err error
		rsp models.Response
	)

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			if msg, err = c.readMessage(); err != nil {
				c.closeHandler(-1, err.Error())

				return
			}

			switch {
			case bytes.Contains(msg, infoPattern):
				if rsp, err = c.handleInfoMsg(msg); err != nil {
					log.Println("Fail to parse info msg:", err, string(msg))
				}

				continue
			case bytes.Contains(msg, subPattern):
				if rsp, err = c.handlSubMsg(msg); err != nil {
					log.Println("Fail to parse subscribe response:", err, string(msg))
				}

				continue
			case bytes.Contains(msg, instrumentPattern):
				if rsp, err = c.handleInsMsg(msg); err != nil {
					log.Println("Fail to parse instrument response:", err, string(msg))
					continue
				}
			case bytes.Contains(msg, tradePattern):
				if rsp, err = c.handleTdMsg(msg); err != nil {
					log.Println("Fail to parse trade response:", err, string(msg))
					continue
				}
			case bytes.Contains(msg, mblPattern):
				if rsp, err = c.handleMblMsg(msg); err != nil {
					log.Println("Fail to parse MBL response:", err, string(msg))
					continue
				}
			case bytes.Contains(msg, authPattern1) || bytes.Contains(msg, authPattern2):
				if rsp, err = c.handleAuthMsg(msg); err != nil {
					log.Println("Fail to parse authentication response:", err, string(msg))
					continue
				}
			default:
				log.Println("Unkonw table type:", string(msg))
			}

			if logLevel >= 2 {
				log.Println("<-", rsp.ToString())
			}
		}
	}
}

func (c *client) generateSignature(
	secret, method string, url *url.URL, expires int, body *bytes.Buffer) string {
	h := hmac.New(sha256.New, []byte(secret))

	path := url.Path
	if url.RawQuery != "" {
		path = path + "?" + url.RawQuery
	}

	var bodyString string
	if body != nil {
		bodyString = strings.TrimRight(body.String(), "\r\n")
	}

	message := strings.ToUpper(method) + path + strconv.Itoa(expires) + bodyString

	// log.Println("message:", message)

	h.Write([]byte(message))

	signature := hex.EncodeToString(h.Sum(nil))

	// log.Println("signature:", signature)
	return signature
}

// NewClient create a new mock client instance
func NewClient(cfg *WsConfig) Client {
	ins := client{
		cfg:           cfg,
		heartbeatChan: make(chan *models.HeartBeat),
		closeFlag:     make(chan bool, 0),

		SubscribedTopics: make(map[string]*models.SubscribeResponse),
	}

	return &ins
}
