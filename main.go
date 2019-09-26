package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/frozenpine/wstester/modules"

	flag "github.com/spf13/pflag"
)

const (
	defaultScheme = "wss"
	defaultHost   = "www.btcmex.com"
	defaultPort   = 443
	defaultURI    = "/realtime"

	defaultHBInterval  = 15
	defaultHBFailCount = 3

	// delay in second
	defaultReconnectDelay = 3
	defaultMaxDelayCount  = 5

	defaultDuration = time.Duration(-1)
)

var (
	scheme string
	host   string
	port   int
	uri    string

	dbgLevel int

	hbInterval  int
	hbFailCount int

	deadline time.Duration

	apiKey    string
	apiSecret string
)

func getURL() string {
	var hostString string

	if port == 80 || port == 443 {
		hostString = host
	} else {
		hostString = fmt.Sprintf("%s:%d", host, port)
	}

	return scheme + "://" + strings.Join([]string{hostString, strings.TrimLeft(uri, "/")}, "/")
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

// algorithm ref: https://en.wikipedia.org/wiki/Exponential_backoff
func expectBackoff(failCount, i int, slot int) time.Duration {
	if failCount > i {
		failCount = i
	}

	N := 1<<uint(failCount) - 1

	return time.Millisecond * time.Duration(int64(slot)*int64(N)*1000/2)
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
		&deadline, "deadline", "d", defaultDuration,
		"Deadline duration, -1 means infinity.")

	flag.StringVar(&apiKey, "key", "", "API Key for order request.")
	flag.StringVar(&apiSecret, "secret", "", "API Secret for order request.")
}

func getContext(deadline time.Duration) (context.Context, context.CancelFunc) {
	ctx, cancelFunc := context.WithCancel(context.Background())

	if deadline > 0 {
		ctx, cancelFunc = context.WithDeadline(ctx, time.Now().Add(deadline))
	}

	return ctx, cancelFunc
}

func main() {
	if !flag.Parsed() {
		flag.Parse()
	}

	modules.SetLogLevel(dbgLevel)

	roundCount := 1
	failCount := 0

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt)

	running := true

	for running {
		if failCount > 0 {
			delay := expectBackoff(failCount, defaultMaxDelayCount, defaultReconnectDelay)
			delay += time.Millisecond * time.Duration(rand.Intn(100))

			log.Println("Reconnect after:", delay)

			select {
			case <-time.After(delay):
			case <-sigChan:
				running = false
				return
			}
		}

		ctx, cancelFunc := getContext(deadline)

		cfg := modules.NewConfig()
		if err := cfg.ChangeHost(getURL()); err != nil {
			log.Println(err)
			return
		}
		cfg.HeartbeatInterval = hbInterval
		cfg.HeartbeatFailCount = hbFailCount

		client := modules.NewClient(cfg)

		start := time.Now()
		if err := client.Connect(ctx, ""); err != nil {
			log.Println(err)

			failCount++

			continue
		}

		failCount = 0

		select {
		case <-ctx.Done():
			log.Println(ctx.Err())
			running = false
		case <-client.Closed():
			// gracefully quit heartbeatHandler and other goroutine
			cancelFunc()

			failCount++
		case <-sigChan:
			running = false
			cancelFunc()
		}

		log.Printf(
			"%s round connection last %v long.",
			humanReadNum(roundCount), time.Now().Sub(start),
		)
		roundCount++
	}
}
