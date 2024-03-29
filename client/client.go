package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/frozenpine/wstester/models"
	"github.com/frozenpine/wstester/utils"
	"github.com/frozenpine/wstester/utils/log"
	"github.com/gorilla/websocket"
)

var (
	cacheMapper = map[string]func(context.Context, string) utils.Cache{
		"orderBookL2":    utils.NewMBLCache,
		"orderBookL2_25": utils.NewMBLCache,
		"trade":          utils.NewTradeCache,
		"instrument":     utils.NewInstrumentCache,
	}
)

// Client client instance
type Client interface {
	Host() string
	Connect(ctx context.Context) error
	Closed() <-chan struct{}
	Subscribe(topics ...string)
	UnSubscribe(topics ...string)
	IsConnected() bool
	IsAuthencated() bool
	SendJSONMessage(msg interface{}) error
	SetInfoHandler(func(*models.InfoResponse))
	SetSubHandler(func(*models.SubscribeResponse))
	SetErrHandler(func(*models.ErrResponse))
	GetResponse(string) <-chan models.TableResponse
}

type client struct {
	cfg         *Config
	ws          *websocket.Conn
	connected   bool
	authencated bool
	ctx         context.Context
	lock        sync.Mutex

	infoHandler func(*models.InfoResponse)
	subHandler  func(*models.SubscribeResponse)
	errHandler  func(*models.ErrResponse)

	heartbeatChan  chan *models.HeartBeat
	heartbeatTimer *time.Timer

	rspCache map[string]utils.Cache

	closeFlag chan struct{}
	closeOnce sync.Once

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

		header["api-signature"] = []string{utils.GenerateSignature(
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
			log.Warn("Invalid topic name: ", topic)
			continue
		}

		if c.isSubscribed(topic) {
			log.Warnf("Topic[%s] already subscirbed.", topic)
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
			log.Warn("Invalid topic name: ", topic)
			continue
		}

		if !c.isSubscribed(topic) {
			log.Warnf("Topic[%s] is not subscribed.", topic)
			continue
		}

		c.SubscribedTopics[topic] = nil

		// TODO: real unsubscribe action
	}
}

func (c *client) createCache(topic string) {
	if _, exist := c.rspCache[topic]; exist {
		return
	}

	c.rspCache[topic] = cacheMapper[topic](c.ctx, c.cfg.Symbol)
}

// Connect to remote host
func (c *client) Connect(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	c.ctx = ctx

	remote := c.cfg.GetURL()

	var subList []string

	for topic := range c.SubscribedTopics {
		if IsPublicTopic(topic) {
			c.createCache(topic)

			subList = append(subList, c.normalizeTopic(topic))

			continue
		}

		if IsPrivateTopic(topic) && c.hasAuth() {
			c.createCache(topic)

			subList = append(subList, c.normalizeTopic(topic))
		}
	}

	if len(subList) > 0 {
		remote.RawQuery = "subscribe=" + strings.Join(subList, ",")
	}

	log.Info("Connecting to: ", remote.String())

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
func (c *client) Closed() <-chan struct{} {
	return c.closeFlag
}

func (c *client) closeHandler(code int, msg string) error {
	c.closeOnce.Do(func() {
		close(c.closeFlag)
	})

	c.connected = false
	c.authencated = false

	log.Infof("Websocket closed with code[%d]: %s", code, msg)

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

// SetErrHandler set error response handler, must be setted before calling Connect
func (c *client) SetErrHandler(fn func(*models.ErrResponse)) {
	if fn != nil {
		c.errHandler = fn
	}
}

// SetSubHandler set subscribe response handler, must be setted before calling Connect & Subscribe
func (c *client) SetSubHandler(fn func(*models.SubscribeResponse)) {
	if fn != nil {
		c.subHandler = fn
	}
}

func (c *client) GetResponse(topic string) <-chan models.TableResponse {
	if _, exist := c.SubscribedTopics[topic]; !exist {
		log.Infof("Topic[%s] not subscribed.", topic)
		return nil
	}

	if cache, exist := c.rspCache[topic]; exist {
		_, ch := cache.GetDefaultChannel().RetriveData()
		return ch
	}

	return nil
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
				case <-time.NewTicker(c.cfg.HeartbeatInterval).C:
					c.heartbeatChan <- models.NewPing()
				}
			}
		}()
	} else {
		c.heartbeatTimer = time.NewTimer(c.cfg.HeartbeatInterval * time.Duration(c.cfg.HeartbeatFailCount))

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
			if c.cfg.ReversHeartbeat {
				if err = c.ws.WriteMessage(websocket.TextMessage, []byte("pong")); err != nil {
					c.closeHandler(-1, "Send heartbeat failed: "+hb.String())
					return
				}

				log.Debug("<- ", hb.String())
				log.Debug("-> ", models.NewPong().String())
			} else {
				if err = c.ws.WriteMessage(websocket.TextMessage, []byte("ping")); err != nil {
					c.closeHandler(-1, "Send heartbeat failed: "+hb.String())
					return
				}

				heartbeatCounter += hb.Value()

				log.Debug("-> ", hb.String())
			}
		case "Pong":
			heartbeatCounter += hb.Value()

			log.Debug("<- ", hb.String())
		default:
			log.Error("Invalid heartbeat type: ", hb.String())

			continue
		}

		if heartbeatCounter >= c.cfg.HeartbeatFailCount || heartbeatCounter < 0 {
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
			c.heartbeatTimer.Reset(c.cfg.HeartbeatInterval * time.Duration(c.cfg.HeartbeatFailCount))
		}

		switch {
		case models.PongPattern.Match(msg):
			c.heartbeatChan <- models.NewPong()
		case models.PingPattern.Match(msg):
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
			log.Info("Info: ", info.String())
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

		log.Info("Auth: ", auth.String())
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
			log.Info("Subscribe: ", sub.String())
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
		if insCache, exist := c.rspCache[insRsp.Table]; exist && insCache != nil {
			if !c.cfg.disableCache {
				insCache.Append(utils.NewCacheInput(&insRsp))
			} else {
				insCache.GetDefaultChannel().PublishData(&insRsp)
			}
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
		if tdCache, exist := c.rspCache[tdRsp.Table]; exist && tdCache != nil {
			if !c.cfg.disableCache {
				tdCache.Append(utils.NewCacheInput(&tdRsp))
			} else {
				tdCache.GetDefaultChannel().PublishData(&tdRsp)
			}
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
		if mblCache, exist := c.rspCache[mblRsp.Table]; exist && mblCache != nil {
			if !c.cfg.disableCache {
				mblCache.Append(utils.NewCacheInput(&mblRsp))
			} else {
				mblCache.GetDefaultChannel().PublishData(&mblRsp)
			}
		}
	}()

	return &mblRsp, nil
}

func (c *client) handleErrMsg(msg []byte) (*models.ErrResponse, error) {
	var errRsp models.ErrResponse

	if err := json.Unmarshal(msg, &errRsp); err != nil {
		return nil, err
	}

	defer func() {
		if c.errHandler != nil {
			c.errHandler(&errRsp)
		} else {
			log.Warn(errRsp.String())
		}
	}()

	return &errRsp, nil
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
			case models.InfoPattern.Match(msg):
				if rsp, err = c.handleInfoMsg(msg); err != nil {
					log.Errorf("Fail to parse info msg: %s, %s", err.Error(), string(msg))
				}

				continue
			case models.SubPattern.Match(msg):
				if rsp, err = c.handlSubMsg(msg); err != nil {
					log.Errorf("Fail to parse subscribe response: %s, %s", err.Error(), string(msg))
				}

				continue
			case models.ErrPattern.Match(msg):
				if rsp, err = c.handleErrMsg(msg); err != nil {
					log.Errorf("Fail to parse error response: %s, %s", err.Error(), string(msg))
				}

				continue
			case models.AuthPattern.Match(msg):
				if rsp, err = c.handleAuthMsg(msg); err != nil {
					log.Errorf("Fail to parse authentication response: %s, %s", err.Error(), string(msg))
					continue
				}
			case models.InstrumentPattern.Match(msg):
				if rsp, err = c.handleInsMsg(msg); err != nil {
					log.Errorf("Fail to parse instrument response: %s, %s", err.Error(), string(msg))
					continue
				}
			case models.MBLPattern.Match(msg):
				if rsp, err = c.handleMblMsg(msg); err != nil {
					log.Error("Fail to parse MBL response:", err, string(msg))
					continue
				}
			case models.TradePattern.Match(msg):
				if rsp, err = c.handleTdMsg(msg); err != nil {
					log.Error("Fail to parse trade response:", err, string(msg))
					continue
				}
			default:
				log.Error("Unkonw response type:", string(msg))
				continue
			}

			if log.IsTraceLevel {
				log.Debug("<- ", rsp.String())
			}
		}
	}
}

// NewClient create a new mock client instance
func NewClient(cfg *Config) Client {
	ins := client{
		cfg:           cfg,
		heartbeatChan: make(chan *models.HeartBeat),
		closeFlag:     make(chan struct{}, 0),

		SubscribedTopics: make(map[string]*models.SubscribeResponse),
		rspCache:         make(map[string]utils.Cache),
	}

	return &ins
}
