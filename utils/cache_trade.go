package utils

import (
	"context"
	"log"
	"sync"

	"github.com/frozenpine/ngerest"
	"github.com/frozenpine/wstester/models"
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

	trimLen := MinInts(c.maxLength, hisLen, depth)

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

	if in.msg == nil {
		log.Println("Trade notify content is empty:", in.msg.String())
		return
	}

	if td, ok := in.msg.(*models.TradeResponse); ok {
		c.applyData(td)

		c.channelGroup[Realtime][0].PublishData(td)
	} else {
		log.Println("Can not convert cache input to TradeResponse:", in.msg.String())
	}
}

func (c *TradeCache) applyData(data *models.TradeResponse) {
	c.historyTrade = append(c.historyTrade, data.Data...)

	if hisLen := len(c.historyTrade); hisLen > c.maxLength*3 {
		trimLen := int(float64(c.maxLength) * 1.5)

		c.historyTrade = c.historyTrade[hisLen-trimLen:]
	}
}

// NewTradeCache make a new trade cache.
func NewTradeCache(ctx context.Context, symbol string) Cache {
	if ctx == nil {
		ctx = context.Background()
	}

	td := TradeCache{}
	td.Symbol = symbol
	td.ctx = ctx
	td.maxLength = defaultTradeLen
	td.handleInputFn = td.handleInput
	td.snapshotFn = td.snapshot
	td.pipeline = make(chan *CacheInput, 1000)
	td.ready = make(chan struct{})
	td.channelGroup[Realtime] = map[int]Channel{
		0: &rspChannel{ctx: ctx, retriveLock: sync.Mutex{}, connectLock: sync.Mutex{}},
	}

	if err := td.Start(); err != nil {
		log.Panicln(err)
	}

	return &td
}
