package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"time"

	"./models"
	"github.com/gorilla/websocket"
	flag "github.com/spf13/pflag"
)

const (
	defaultScheme = "wss"
	defaultHost   = "www.btcmex.com"
	defaultPort   = 443
	defaultURI    = "/realtime"

	defaultHBInterval  = 30
	defaultHBFailCount = 3

	defaultDuration = time.Duration(-1)
)

var (
	running = true

	scheme string
	host   string
	port   int
	uri    string

	dbgLevel int

	hbInterval  int
	hbFailCount int

	duration time.Duration

	apiKey    string
	apiSecret string

	pingMsg       = []byte("ping")
	pongMsg       = []byte("pong")
	infoMsg       = []byte(`"info"`)
	instrumentMsg = []byte(`"instrument"`)
	tradeMsg      = []byte(`"trade"`)
	mblMsg        = []byte(`"orderBook`)
	subMsg        = []byte(`"subscribe"`)
)

func infoHandler(info *models.InfoResponse) {
	data, _ := json.Marshal(info)
	log.Println("Info:", string(data))
}

func subscribeHandler(sub *models.SubscribeResponse) {
	rsp, _ := json.Marshal(sub)
	log.Println("Subscribe:", string(rsp))
}

func instrumentHandler(rsp *models.InstrumentResponse) {
	// for _, data := range rsp.Data {
	// 	if data.Symbol == "XBTUSD" && data.LastPrice <= 0 {
	// 		ins, _ := json.Marshal(data)

	// 		fmt.Println("Instrument", rsp.Action, string(ins))
	// 	}
	// }
}

func tradeHandler(rsp *models.TradeResponse) {
	// for _, data := range rsp.Data {
	// 	td, _ := json.Marshal(data)

	// 	fmt.Println("Trade", rsp.Action, string(td))
	// }
}

func mblHandler(rsp *models.MBLResponse) {
	// for _, data := range rsp.Data {
	// 	mbl, _ := json.Marshal(data)

	// 	fmt.Println("MBL", rsp.Action, string(mbl))
	// }
}

func wsMessageHandler(
	done chan<- bool, hb chan<- *models.Heartbeat, ws *websocket.Conn) {
	defer close(done)

	for running {
		_, msg, err := ws.ReadMessage()

		if err != nil && running {
			log.Println("Error:", err)

			return
		}

		switch {
		case bytes.Contains(msg, pongMsg):
			hb <- models.NewPong()
		case bytes.Contains(msg, infoMsg):
			var info models.InfoResponse

			if err = json.Unmarshal(msg, &info); err != nil {
				log.Println("Fail to parse info msg:", err, string(msg))
			} else {
				infoHandler(&info)
			}
		case bytes.Contains(msg, subMsg):
			var sub models.SubscribeResponse

			if err = json.Unmarshal(msg, &sub); err != nil {
				log.Println(
					"Fail to parse subscribe response:", err, string(msg))
			} else {
				subscribeHandler(&sub)
			}
		case bytes.Contains(msg, instrumentMsg):
			var insRsp models.InstrumentResponse

			if err = json.Unmarshal(msg, &insRsp); err != nil {
				log.Println(
					"Fail to parse instrument response:", err, string(msg))
			} else {
				instrumentHandler(&insRsp)
			}
		case bytes.Contains(msg, tradeMsg):
			var tdRsp models.TradeResponse

			if err = json.Unmarshal(msg, &tdRsp); err != nil {
				log.Println("Fail to parse trade response:", err, string(msg))
			} else {
				tradeHandler(&tdRsp)
			}
		case bytes.Contains(msg, mblMsg):
			var mblRsp models.MBLResponse

			if err = json.Unmarshal(msg, &mblRsp); err != nil {
				log.Println("Fail to parse MBL response:", err, string(msg))
			} else {
				mblHandler(&mblRsp)
			}
		default:
			if len(msg) > 0 {
				log.Println("Unkonw table type:", string(msg))
			}
		}

		if dbgLevel > 1 {
			log.Println("  <", string(msg))
		}
	}
}

func heartbeatHandler(hbChan <-chan *models.Heartbeat, ws *websocket.Conn) {
	var heartbeatCounter int

	for hb := range hbChan {
		switch hb.Type() {
		case "Ping":
			err := ws.WriteMessage(websocket.TextMessage, []byte("ping"))

			if err != nil && running {
				log.Println("Send heartbeat failed:", err)

				ws.Close()

				return
			}

			// if dbgLevel > 0 {
			log.Println(">  ", hb.ToString())
			// }
		case "Pong":
			// if dbgLevel > 0 {
			log.Println("  <", hb.ToString())
			// }
		}

		heartbeatCounter += hb.Value()

		if int(heartbeatCounter) > hbFailCount || heartbeatCounter < 0 {
			log.Println("Heartbeat miss-match:", heartbeatCounter)

			ws.Close()

			return
		}
	}
}

func hostString() string {
	if port == 80 || port == 443 {
		return host
	}

	return fmt.Sprintf("%s:%d", host, port)
}

func connect(ctx context.Context) (*websocket.Conn, error) {
	remote := url.URL{
		Scheme:   scheme,
		Host:     hostString(),
		Path:     uri,
		RawQuery: "subscribe=instrument:XBTUSD,orderBookL2:XBTUSD,trade:XBTUSD",
	}

	log.Println("Connect to:", remote.String())

	c, rsp, err := websocket.DefaultDialer.DialContext(
		ctx, remote.String(), nil)

	if err != nil {
		log.Printf("Fail to connect[%s]: %v\n%s",
			remote.String(), err, rsp.Status)

		return nil, err
	}

	return c, nil
}

func testRound(ctx context.Context, count int, deadline <-chan struct{}) error {
	c, err := connect(ctx)
	if err != nil {
		return err
	}

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt)

	start := time.Now()
	defer c.Close()

	done := make(chan bool, 1)
	hbChan := make(chan *models.Heartbeat)

	go wsMessageHandler(done, hbChan, c)
	go heartbeatHandler(hbChan, c)

	ticker := time.NewTicker(time.Second * time.Duration(hbInterval))
	defer ticker.Stop()

	for {
		select {
		case <-done:
			lastLong := time.Now().Sub(start)

			log.Printf(
				"%s round connection last %v long.",
				humanReadNum(count), lastLong,
			)

			close(hbChan)

			return nil
		case <-ticker.C:
			hbChan <- models.NewPing()
		case <-sigChan:
			log.Println("Keyboard interupt, waiting for exit...")
			running = false
			c.Close()
		case <-deadline:
			log.Printf("Deadline %v exceeded.", duration)
			running = false
			c.Close()
		}
	}
}

func humanReadNum(num int) string {
	switch num {
	case 1:
		return strconv.Itoa(num) + "st"
	case 2:
		return strconv.Itoa(num) + "nd"
	case 3:
		return strconv.Itoa(num) + "rd"
	default:
		return strconv.Itoa(num) + "th"
	}
}

func init() {
	flag.StringVar(&scheme, "scheme", defaultScheme, "Websocket scheme.")
	flag.StringVarP(
		&host, "host", "H", defaultHost, "Host addreses to connect.")
	flag.IntVarP(
		&port, "port", "p", defaultPort, "Host port to connect.")
	flag.StringVar(&uri, "uri", defaultURI, "URI for realtime push data.")

	flag.CountVarP(
		&dbgLevel, "verbose", "v",
		"Debug level, turn on for detail info.")

	flag.IntVar(
		&hbInterval, "heartbeat", defaultHBInterval,
		"Heartbeat interval in seconds.")
	flag.IntVar(
		&hbFailCount, "fail", defaultHBFailCount,
		"Heartbeat fail count.")

	flag.DurationVarP(
		&duration, "duration", "d", defaultDuration,
		"Deadline duration, -1 means infinity.")

	flag.StringVar(&apiKey, "key", "", "API Key for order request.")
	flag.StringVar(&apiSecret, "secret", "", "API Secret for order request.")
}

func main() {
	if !flag.Parsed() {
		flag.Parse()
	}

	roundCount := 1

	deadline := make(chan struct{})

	ctx, cancelFunc := context.WithCancel(context.Background())

	if duration > 0 {
		go func() {
			<-time.After(duration)
			close(deadline)
		}()
	}

	for running {
		select {
		case <-deadline:
			log.Printf("Deadline %v exceeded.", duration)
			running = false
			cancelFunc()
			return
		default:
			log.Printf("Starting %s round test...", humanReadNum(roundCount))

			err := testRound(ctx, roundCount, deadline)

			if err != nil {
				log.Fatalln(err)
				return
			}

			if running {
				roundCount++
				<-time.After(time.Second * 3)
			} else {
				return
			}
		}
	}
}
