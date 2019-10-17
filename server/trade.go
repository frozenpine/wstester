package server

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/frozenpine/ngerest"
	"github.com/frozenpine/wstester/models"
	"github.com/frozenpine/wstester/utils"
)

var (
	defaultTradeLen int = 200
)

// TradeCache retrive & store trade data
type TradeCache struct {
	tableCache

	historyTrade []*ngerest.Trade
}

func (c *TradeCache) snapshot() models.TableResponse {
	snap := models.NewTradePartial()

	hisLen := len(c.historyTrade)

	trimLen := utils.MinInt(c.maxLength, hisLen)

	snap.Data = c.historyTrade[hisLen-trimLen:]

	return snap
}

func (c *TradeCache) handleInput(in *CacheInput) models.TableResponse {
	var rsp models.TableResponse

	if in.IsBreakPoint() {
		rsp = in.breakpointFunc()
	} else {
		// FIXME: real sub flow handle
		td := models.TradeResponse{}
		td.Table = "trade"
		td.Action = "insert"

		parsed := ngerest.Trade{}

		json.Unmarshal(in.msg, &parsed)

		td.Data = []*ngerest.Trade{&parsed}

		c.applyData(&td)

		rsp = &td
	}

	return rsp
}

func (c *TradeCache) applyData(data *models.TradeResponse) {
	c.historyTrade = append(c.historyTrade, data.Data...)

	if hisLen := len(c.historyTrade); hisLen > c.maxLength*3 {
		trimLen := int(float64(c.maxLength) * 1.5)

		c.historyTrade = c.historyTrade[hisLen-trimLen:]
	}
}

func mockTrade(cache Cache) {
	go func() {
		ticker := time.NewTicker(time.Second)

		var (
			lastPrice         float64
			lastTickDirection string
			sides                     = [2]string{"Buy", "Sell"}
			sizeMax                   = float32(1000.0)
			priceMax                  = 9900.0
			priceMin                  = 8100.0
			hisMaxRate        float64 = 0.0
		)

		for {
			<-ticker.C

			start := time.Now()
			rand.Seed(start.UnixNano())
			count := rand.Intn(1000)

			for i := 0; i < count; i++ {

				choice := rand.Intn(1000)

				tickDirection := ""

				price := priceMin + float64(choice)*0.5
				if price > priceMax {
					price = priceMax
				}
				switch {
				case price > lastPrice:
					tickDirection = "PlusTick"
				case price == lastPrice:
					if strings.Contains(lastTickDirection, "Zero") {
						tickDirection = lastTickDirection
					} else {
						tickDirection = "Zero" + lastTickDirection
					}
				case price < lastPrice:
					tickDirection = "MinusTick"
				}

				lastTickDirection = tickDirection

				size := float32(choice)
				if size < 1 {
					size = 1
				} else if size > sizeMax {
					size = sizeMax
				}

				lastPrice = price

				td := ngerest.Trade{
					Symbol:        "XBTUSD",
					Side:          sides[choice%2],
					Size:          size,
					Price:         price,
					Timestamp:     ngerest.NGETime(time.Now()),
					TickDirection: tickDirection,
				}

				lastPrice = td.Price

				result, _ := json.Marshal(td)

				cache.Append(NewCacheInput(result))
			}

			elasped := time.Now().Sub(start).Nanoseconds()
			rate := float64(count) * 1000.0 / float64(elasped/1000)
			if rate > hisMaxRate && !math.IsInf(rate, 1) {
				hisMaxRate = rate
			}

			log.Printf("Mock trade send rate: %.2f rps, history max rate: %.2f rps\n", rate, hisMaxRate)
		}
	}()
}

// NewTradeCache make a new trade cache.
func NewTradeCache(ctx context.Context) *TradeCache {
	td := TradeCache{}

	td.pipeline = make(chan *CacheInput, 1000)
	td.destinations = make(map[Session]chan models.Response)
	td.ready = make(chan struct{})
	td.maxLength = defaultTradeLen
	td.handleInputFn = td.handleInput
	td.snapshotFn = td.snapshot

	if err := td.Start(ctx); err != nil {
		log.Panicln(err)
	}

	return &td
}
