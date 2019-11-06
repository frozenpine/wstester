package utils

import (
	"context"

	"github.com/frozenpine/ngerest"
	"github.com/frozenpine/wstester/models"
	"github.com/frozenpine/wstester/utils/log"
)

const (
	maxInsLength int = (3600 / 5) * 24
)

// WAP weighted average price for both side
type WAP struct {
	Buy, Sell float64
}

// InstrumentCache retrive & store instrument data
type InstrumentCache struct {
	tableCache

	wapPriceList        []*WAP
	indicativePriceList []float64
	markPriceList       []float64

	instrument *ngerest.Instrument
}

func (c *InstrumentCache) snapshot(depth int) models.TableResponse {
	rsp := models.NewInstrumentPartial()

	rsp.Data = append(rsp.Data, c.instrument)

	return rsp
}

func (c *InstrumentCache) handleInput(input *CacheInput) {
	if c.handleBreakpoint(input) {
		return
	}

	if ins, ok := input.msg.(*models.InstrumentResponse); ok {
		if c.applyData(ins) {
			c.channelGroup[Realtime][0].PublishData(ins)
		}
	} else {
		log.Error("Can not convert cache input to InstrumentResponse: ", input.msg.String())
	}
}

func (c *InstrumentCache) applyInsPrice(indPrice, markPrice float64) {
	c.indicativePriceList = append(c.indicativePriceList, indPrice)
	c.markPriceList = append(c.markPriceList, markPrice)

	if length := len(c.indicativePriceList); length > maxInsLength*maxMultiple {
		c.indicativePriceList = c.indicativePriceList[length-maxInsLength*maxMultiple/2:]
	}

	if length := len(c.markPriceList); length > maxInsLength*maxMultiple {
		c.markPriceList = c.markPriceList[length-maxInsLength*maxMultiple/2:]
	}
}

func (c *InstrumentCache) applyData(ins *models.InstrumentResponse) bool {
	data := ins.Data[0]

	changed := false

	switch ins.Action {
	case models.PartialAction:
		c.instrument = data
	case models.UpdateAction:
		if data.IndicativeSettlePrice > 0 && data.MarkPrice > 0 {
			c.applyInsPrice(data.IndicativeSettlePrice, data.MarkPrice)

			changed = true
		}

		if data.BidPrice > 0 {
			if c.instrument.BidPrice != data.BidPrice {
				c.instrument.BidPrice = data.BidPrice

				changed = true
			}

			if c.instrument.AskPrice != data.AskPrice {
				c.instrument.AskPrice = data.AskPrice

				changed = true
			}
		}

		if data.LastPrice > 0 {
			c.instrument.LastPrice = data.LastPrice

			c.instrument.Volume = data.Volume
			c.instrument.Volume24h = data.Volume24h
			c.instrument.TotalVolume = data.TotalVolume
			c.instrument.PrevTotalVolume = data.PrevTotalVolume

			c.instrument.Turnover = data.Turnover
			c.instrument.Turnover24h = data.Turnover24h
			c.instrument.TotalTurnover = data.TotalTurnover
			c.instrument.PrevTotalTurnover = data.PrevTotalTurnover

			changed = true
		}

		if data.FundingRate > 0 {
			c.instrument.FundingRate = data.FundingRate
			c.instrument.FundingTimestamp = data.FundingTimestamp

			changed = true
		}
	default:
		log.Error("Invalid action for instrument cache: ", ins.Action)
	}

	return changed
}

// NewInstrumentCache make a new instrument cache.
func NewInstrumentCache(ctx context.Context, symbol string) Cache {
	if ctx == nil {
		ctx = context.Background()
	}

	ins := InstrumentCache{}
	ins.Symbol = symbol
	ins.ctx = ctx
	ins.handleInputFn = ins.handleInput
	ins.snapshotFn = ins.snapshot
	ins.pipeline = make(chan *CacheInput, 1000)
	ins.ready = make(chan struct{})
	ins.channelGroup[Realtime] = map[int]Channel{
		0: &rspChannel{
			ctx:           ctx,
			destinations:  map[string]chan<- models.TableResponse{},
			childChannels: map[string]Channel{},
		},
	}

	if err := ins.Start(); err != nil {
		log.Panic(err)
	}

	return &ins
}
