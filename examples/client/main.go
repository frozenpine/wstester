package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/frozenpine/wstester/client"
	"github.com/frozenpine/wstester/utils"
	"github.com/frozenpine/wstester/utils/log"

	flag "github.com/spf13/pflag"
)

const (
	defaultSymbol = "XBTUSD"
	defaultScheme = "wss"
	defaultHost   = "www.btcmex.com"
	defaultPort   = 443
	defaultURI    = "/realtime"

	defaultHBInterval  = time.Second * 15
	defaultHBFailCount = 3

	// delay in second
	defaultReconnectDelay    = 3
	defaultMaxReconnectCount = -1
	defaultMaxDelayCount     = 6

	defaultRunningDuration = time.Duration(0)
)

var (
	symbol string
	scheme string
	host   string
	port   int
	uriStr string
	urlStr string

	defaultTopics = []string{"trade", "orderBookL2", "instrument"}
	topics        []string
	appendTopics  bool

	dbgLevel int

	hbInterval  time.Duration
	hbFailCount int

	reconnectDelay    int
	maxReconnectCount int
	maxDelayCount     int

	deadline time.Duration

	apiKey    string
	apiSecret string

	cacheMap = make(map[string]utils.Cache)
)

func getURL() string {
	if urlStr != "" {
		if !strings.HasPrefix(urlStr, "http") && !strings.HasPrefix("urlStr", "ws") {
			urlStr = defaultScheme + "://" + urlStr
		}

		parsed, err := url.Parse(urlStr)
		if err != nil {
			log.Panic(err)
		}

		scheme = parsed.Scheme
		host = parsed.Hostname()

		switch parsed.Port() {
		case "":
			switch scheme {
			case "http":
				fallthrough
			case "ws":
				port = 80
			case "https":
				fallthrough
			case "wss":
				port = 443
			default:
				log.Panic("unsupported scheme: ", scheme)
			}
		default:
			port, _ = strconv.Atoi(parsed.Port())
		}

		if parsed.Path != "" {
			uriStr = parsed.Path
		}
	}

	var hostString string

	if port == 80 || port == 443 {
		hostString = host
	} else {
		hostString = fmt.Sprintf("%s:%d", host, port)
	}

	return strings.Replace(scheme, "http", "ws", 1) + "://" + strings.Join([]string{hostString, strings.TrimLeft(uriStr, "/")}, "/")
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
func expectBackoff(c, i int, slot int) time.Duration {
	if c > i {
		c = i
	}

	N := 1<<uint(c) - 1

	return time.Millisecond * time.Duration(int64(slot)*int64(N)*1000/2)
}

func init() {
	flag.StringVar(&symbol, "symbol", defaultSymbol, "Symbol name.")
	flag.StringVar(&scheme, "scheme", defaultScheme, "Websocket scheme.")
	flag.StringVarP(
		&host, "host", "H", defaultHost, "Host addreses to connect.")
	flag.IntVarP(
		&port, "port", "p", defaultPort, "Host port to connect.")
	flag.StringVar(&uriStr, "uri", defaultURI, "URI for realtime push data.")
	flag.StringVar(&urlStr, "url", "", "Connection URL.")

	flag.StringSliceVar(
		&topics, "topics", defaultTopics,
		"Topic names for subscribe.")
	flag.BoolVar(&appendTopics, "append", false,
		"Wether append topic list to default subscrib.")

	flag.CountVarP(
		&dbgLevel, "verbose", "v",
		"Debug level, turn on for detail info.")

	flag.DurationVar(
		&hbInterval, "heartbeat", defaultHBInterval,
		"Heartbeat interval.")
	flag.IntVar(
		&hbFailCount, "fail", defaultHBFailCount,
		"Heartbeat fail count.")

	flag.IntVar(
		&reconnectDelay, "delay", defaultReconnectDelay,
		"Delay seconds per binary expect backoff algorithm's delay slot.")
	flag.IntVar(
		&maxReconnectCount, "max-retry", defaultMaxReconnectCount,
		"Max reconnect count, -1 means infinity.")
	flag.IntVar(
		&maxDelayCount, "max-count", defaultMaxDelayCount,
		"Max slot count in binary expect backoff algorithm.")

	flag.DurationVarP(
		&deadline, "deadline", "d", defaultRunningDuration,
		"Deadline duration, must be positive to take effect.")

	flag.StringVar(&apiKey, "key", "", "API Key for authentication request.")
	flag.StringVar(&apiSecret, "secret", "", "API Secret for authentication request.")

	// log.SetFlags(log.Lmicroseconds | log.Ldate)
}

func getContext(deadline time.Duration) (context.Context, context.CancelFunc) {
	ctx, cancelFunc := context.WithCancel(context.Background())

	if deadline > 0 {
		ctx, cancelFunc = context.WithDeadline(ctx, time.Now().Add(deadline))
	}

	if apiKey != "" && apiSecret != "" {
		ctx = context.WithValue(
			ctx, client.ContextAPIKey,
			client.APIKeyAuth{
				Key:     apiKey,
				Secret:  apiSecret,
				AuthURI: "/api/v1/signature",
			})
	}

	return ctx, cancelFunc
}

func normalizeTopicTable() (err error) {
	if appendTopics {
		topics = append(topics, defaultTopics...)
	}
	topicSet := utils.NewStringSet(topics)
	topics = topicSet.Values()

	for _, topic := range topics {
		utils.RegisterTableModel(topic, topicMapper[topic])
	}

	filters, err = utils.ParseSQL(sql)

	return
}

func main() {
	if !flag.Parsed() {
		flag.Parse()
	}

	if dbgLevel > 1 {
		log.SetLogLevel(log.TraceLevel)
	} else if dbgLevel > 0 {
		log.SetLogLevel(log.DebugLevel)
	}

	if err := normalizeTopicTable(); err != nil {
		log.Fatal(err)
	}

	roundCount := 1
	failCount := 0

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	running := true

	maxLast := time.Duration(0)

	progStart := time.Now()

	for running {
		if failCount > 0 {
			if maxReconnectCount >= 0 && failCount > maxReconnectCount {
				running = false
				return
			}

			delay := expectBackoff(failCount, maxDelayCount, reconnectDelay)
			delay += time.Millisecond * time.Duration(rand.Intn(100))

			log.Warn("Reconnect after: ", delay)

			select {
			case <-time.After(delay):
			case <-sigChan:
				running = false
				return
			}
		}

		ctx, cancelFunc := getContext(deadline)

		cfg := client.NewConfig()
		if err := cfg.ChangeHost(getURL()); err != nil {
			log.Error(err)
			return
		}
		cfg.HeartbeatInterval = hbInterval
		cfg.HeartbeatFailCount = hbFailCount
		cfg.Symbol = symbol

		ins := client.NewClient(cfg)
		ins.Subscribe(topics...)

		start := time.Now()
		if err := ins.Connect(ctx); err != nil {
			log.Error(err)

			failCount++

			continue
		}

		for tableName := range filters {
			filter(ctx, tableName, ins.GetResponse(tableName))
		}

		failCount = 0

		select {
		case <-ctx.Done():
			log.Info(ctx.Err())
			running = false
		case <-ins.Closed():
			// gracefully quit heartbeatHandler and other goroutine
			cancelFunc()
			failCount++
		case <-sigChan:
			cancelFunc()
			running = false
		}

		last := time.Now().Sub(start)
		if last > maxLast {
			maxLast = last
		}
		log.Infof(
			"Program starts at [%v], %s round connection last %v long, max connection time in history is %v.",
			progStart, humanReadNum(roundCount), last, maxLast,
		)
		roundCount++
	}
}
