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
	"github.com/frozenpine/wstester/kafka"
	"github.com/frozenpine/wstester/models"
	"github.com/frozenpine/wstester/utils"
)

const (
	defaultTradeLen int = 200
)

// TradeCache retrive & store trade data
type TradeCache struct {
	tableCache

	historyTrade []*ngerest.Trade
}

func (c *TradeCache) snapshot(depth int) models.TableResponse {
	if depth < 1 {
		depth = c.maxLength
	}

	snap := models.NewTradePartial()

	hisLen := len(c.historyTrade)

	trimLen := utils.MinInts(c.maxLength, hisLen, depth)

	snap.Data = c.historyTrade[hisLen-trimLen:]

	return snap
}

func (c *TradeCache) handleInput(in *CacheInput) {
	if in.IsBreakPoint() {
		if rsp := in.breakpointFunc(); rsp != nil {
			if in.pubChannels != nil && len(in.pubChannels) > 0 {
				for _, ch := range in.pubChannels {
					ch.PublishData(rsp)
				}
			}
		}

		return
	}

	tdNotify := kafka.TradeNotify{}

	if err := json.Unmarshal(in.msg, &tdNotify); err != nil {
		log.Println(err)
		return
	}

	if tdNotify.Content == nil {
		log.Println("Trade notify content is empty:", string(in.msg))
		return
	}

	c.applyData(tdNotify.Content)

	c.channelGroup[Realtime][0].PublishData(tdNotify.Content)
}

func (c *TradeCache) applyData(data *models.TradeResponse) {
	c.historyTrade = append(c.historyTrade, data.Data...)

	if hisLen := len(c.historyTrade); hisLen > c.maxLength*3 {
		trimLen := int(float64(c.maxLength) * 1.5)

		c.historyTrade = c.historyTrade[hisLen-trimLen:]
	}
}

func mockTrade(cache Cache) {
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

		tdNotify := kafka.TradeNotify{
			Content: &models.TradeResponse{},
		}
		tdNotify.Type = "trade"
		tdNotify.Content.Table = "trade"
		tdNotify.Content.Action = "insert"

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

			tdNotify.Content.Data = append(tdNotify.Content.Data, &td)
		}

		result, _ := json.Marshal(tdNotify)
		cache.Append(NewCacheInput(result))

		elasped := time.Now().Sub(start).Nanoseconds()
		rate := float64(count) * 1000.0 / float64(elasped/1000)
		if rate > hisMaxRate && !math.IsInf(rate, 1) {
			hisMaxRate = rate
		}

		log.Printf("Mock trade send rate: %.2f rps, history max rate: %.2f rps\n", rate, hisMaxRate)
	}
}

// NewTradeCache make a new trade cache.
func NewTradeCache(ctx context.Context) *TradeCache {
	td := TradeCache{}
	td.ctx = ctx
	td.maxLength = defaultTradeLen
	td.handleInputFn = td.handleInput
	td.snapshotFn = td.snapshot

	if err := td.Start(); err != nil {
		log.Panicln(err)
	}

	return &td
}
