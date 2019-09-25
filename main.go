package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/frozenpine/wstester/mock"

	flag "github.com/spf13/pflag"
)

const (
	defaultScheme = "wss"
	defaultHost   = "www.btcmex.com"
	defaultPort   = 443
	defaultURI    = "/realtime"

	defaultHBInterval  = 15
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
)

func hostString() string {
	if port == 80 || port == 443 {
		return host
	}

	return fmt.Sprintf("%s:%d", host, port)
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

func getContext() (context.Context, context.CancelFunc) {
	ctx, cancelFunc := context.WithCancel(context.Background())

	if duration > 0 {
		ctx, cancelFunc = context.WithDeadline(ctx, time.Now().Add(duration))
	}

	return ctx, cancelFunc
}

func main() {
	if !flag.Parsed() {
		flag.Parse()
	}

	mock.SetLogLevel(dbgLevel)

	roundCount := 1

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt)

	for running {
		ctx, cancelFunc := getContext()

		cfg := mock.NewConfig()
		cfg.ChangeHost(hostString())

		client := mock.NewClient(cfg)

		start := time.Now()
		if err := client.Connect(ctx, ""); err != nil {
			log.Println(err)
		} else {
			select {
			case <-ctx.Done():
				log.Println(ctx.Err())
				running = false
			case <-client.Closed():
				// gracefully quit heartbeatHandler and other goroutine
				cancelFunc()
				// TODO: 指数回退 + 随机延迟 以实现重连延时
			case <-sigChan:
				running = false
				cancelFunc()
			}
		}

		log.Printf(
			"%s round connection last %v long.",
			humanReadNum(roundCount), time.Now().Sub(start),
		)
		roundCount++
	}
}
