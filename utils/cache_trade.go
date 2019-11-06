package utils

import (
	"context"

	"github.com/frozenpine/ngerest"
	"github.com/frozenpine/wstester/models"
	"github.com/frozenpine/wstester/utils/log"
)

const (
	maxTradeLen int = 200
)

// TradeCache retrive & store trade data
type TradeCache struct {
	tableCache

	historyTrade []*ngerest.Trade
}

func (c *TradeCache) snapshot(depth int) models.TableResponse {
	snap := models.NewTradePartial()

	hisLen := len(c.historyTrade)

	var trimLen int
	if depth < 1 {
		trimLen = MinInt(hisLen, maxTradeLen)
	} else {
		trimLen = MinInts(maxTradeLen, hisLen, depth)
	}

	snap.Data = c.historyTrade[hisLen-trimLen:]

	return snap
}

func (c *TradeCache) handleInput(input *CacheInput) {
	if input.IsBreakPoint() {
		c.handleBreakpoint(input)

		return
	}

	if input.msg == nil {
		log.Error("Trade notify content is empty: ", input.msg.String())
		return
	}

	if td, ok := input.msg.(*models.TradeResponse); ok {
		if c.applyData(td) {
			c.channelGroup[Realtime][0].PublishData(td)
		}
	} else {
		log.Error("Can not convert cache input to TradeResponse: ", input.msg.String())
	}
}

func (c *TradeCache) applyData(data *models.TradeResponse) bool {
	publish := false

	switch data.Action {
	case models.PartialAction:
		if len(c.historyTrade) < 1 {
			// 防止client端使用cache时，partial数据无输出的问题
			publish = true
		}

		c.historyTrade = data.Data
	case models.InsertAction:
		publish = true

		c.historyTrade = append(c.historyTrade, data.Data...)

		if hisLen := len(c.historyTrade); hisLen > maxTradeLen*maxMultiple {
			c.historyTrade = c.historyTrade[hisLen-maxTradeLen*maxMultiple/2:]
		}
	default:
		log.Error("Invalid action for trade cache: ", data.Action)
	}

	return publish
}

// NewTradeCache make a new trade cache.
func NewTradeCache(ctx context.Context, symbol string) Cache {
	if ctx == nil {
		ctx = context.Background()
	}

	td := TradeCache{}
	td.Symbol = symbol
	td.ctx = ctx
	td.handleInputFn = td.handleInput
	td.snapshotFn = td.snapshot
	td.pipeline = make(chan *CacheInput, 1000)
	td.ready = make(chan struct{})
	td.channelGroup[Realtime] = map[int]Channel{
		0: &rspChannel{
			ctx:           ctx,
			destinations:  map[string]chan<- models.TableResponse{},
			childChannels: map[string]Channel{},
		},
	}

	if err := td.Start(); err != nil {
		log.Panic(err)
	}

	return &td
}
