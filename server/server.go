package server

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Shopify/sarama"
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
	// RunForever startup and serve forever
	RunForever(ctx context.Context) error
	// ReloadCfg reload server config
	ReloadCfg(*WsConfig)
}

type server struct {
	cfg      *WsConfig
	ctx      context.Context
	upgrader *websocket.Upgrader

	statics serverStatics

	clients     map[string]Session
	subCaches   map[string]sarama.ConsumerGroupHandler
	pubChannels map[string]Channel
}

func (s *server) ReloadCfg(cfg *WsConfig) {
	s.cfg = cfg

	for _, client := range s.clients {
		client.ReloadCfg()
	}
}

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

func (s *server) getReqAuth(r *http.Request) error {
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

func (s *server) getReqSubscribe(r *http.Request) *models.OperationRequest {
	query := r.URL.Query()

	if topicStrList, exist := query["subscribe"]; exist {
		op := models.OperationRequest{}
		op.Operation = "subscribe"

		for _, topicStr := range topicStrList {
			op.Args = append(op.Args, strings.Split(topicStr, ",")...)
		}

		log.Println("Request subscirbe:", op.String())

		return &op
	}

	return nil
}

func (s *server) parseOperation(msg []byte) (*models.OperationRequest, error) {
	req := models.OperationRequest{}

	if err := json.Unmarshal(msg, &req); err != nil {
		return nil, err
	}

	return &req, nil
}

func (s *server) handleAuth(req models.Request, client Session) models.Response {
	return nil
}

func (s *server) handleSubscribe(req models.Request, client Session) []models.Response {
	var (
		rspList    []models.Response
		pubChannel Channel
		chanExist  bool
	)

	for _, topicStr := range req.GetArgs() {
		parsed := strings.Split(topicStr, ":")
		topicName := parsed[0]

		waitRsp := make(chan bool, 0)

		// TODO: private flow subscribe

		if pubChannel, chanExist = s.pubChannels[topicName]; chanExist {
			go func() {
				<-waitRsp

				for data := range pubChannel.RetriveData() {
					client.WriteJSONMessage(data)
				}
			}()
		}

		rsp := models.SubscribeResponse{
			Success:   chanExist,
			Subscribe: topicStr,
			Request:   *req.(*models.OperationRequest),
		}

		rspList = append(rspList, &rsp)
		client.WriteJSONMessage(&rsp)

		close(waitRsp)
	}

	return rspList
}

func (s *server) wsUpgrader(w http.ResponseWriter, r *http.Request) {
	var (
		conn *websocket.Conn
		err  error
	)

	if err = s.getReqAuth(r); err != nil {
		log.Println(err)
		return
	}

	conn, err = s.upgrader.Upgrade(w, r, w.Header())

	if err != nil {
		http.Error(w, err.Error(), 400)
	}

	clientSenssion := s.incClients(conn)
	defer func() {
		s.decClients(clientSenssion)
	}()

	var (
		msg     []byte
		req     models.Request
		rspList []models.Response
	)

	if headerSub := s.getReqSubscribe(r); headerSub != nil {
		if subRsp := s.handleSubscribe(headerSub, clientSenssion); subRsp != nil && logLevel >= 2 {
			for _, rsp := range subRsp {
				log.Println("->", clientSenssion.GetID(), rsp.String())
			}
		}
	}

	for {
		select {
		case <-s.ctx.Done():
			clientSenssion.Close(0, "Server exit.")
			return
		default:
			if msg, err = clientSenssion.ReadMessage(); err != nil {
				clientSenssion.Close(-1, err.Error())
				return
			}

			switch {
			case bytes.Contains(msg, opPattern):
				if req, err = s.parseOperation(msg); err != nil {
					log.Println("Fail to parse request operation:", err, string(msg))
					continue
				}
			default:
				log.Println("Unknow request:", string(msg))
				continue
			}

			switch req.GetOperation() {
			case "subscribe":
				if subRsp := s.handleSubscribe(req, clientSenssion); subRsp != nil {
					rspList = append(rspList, subRsp...)
				}
			case "auth":
				if authRsp := s.handleAuth(req, clientSenssion); authRsp != nil {
					rspList = append(rspList, authRsp)
				}
			default:
				log.Println("Unkown request operation:", req.String())
				continue
			}
		}

		if logLevel >= 2 {
			log.Println("<-", clientSenssion.GetID(), req.String())

			for _, rsp := range rspList {
				log.Println("->", clientSenssion.GetID(), rsp.String())
			}
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
		statics:     serverStatics{},
		clients:     make(map[string]Session),
		subCaches:   make(map[string]sarama.ConsumerGroupHandler),
		pubChannels: make(map[string]Channel),
	}

	td := NewTradeCache()
	ins := NewInstrumentCache()
	mbl := NewMBLCache()

	svr.pubChannels["trade"] = td
	svr.pubChannels["instrument"] = ins
	svr.pubChannels["orderBookL2"] = mbl

	svr.subCaches["trade"] = td
	// svr.subCaches["instrument"] = ins
	// svr.subCaches["orderBookL2"] = mbl

	return &svr
}
