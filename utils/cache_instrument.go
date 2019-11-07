package utils

import (
	"context"
	"encoding/json"
	"errors"

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

	insCache map[string]*ngerest.Instrument
}

func (c *InstrumentCache) snapshot(depth int) models.TableResponse {
	rsp := models.NewInstrumentPartial()

	for _, ins := range c.insCache {
		rsp.Data = append(rsp.Data, ins)
	}

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

func (c *InstrumentCache) initCache() {
	c.insCache = make(map[string]*ngerest.Instrument)
}

func (c *InstrumentCache) handlePartial(rsp *models.InstrumentResponse) bool {
	changed := false

	if c.insCache == nil {
		c.initCache()

		// 防止client端使用cache时，partial数据无输出的问题
		changed = true
	}

	for _, ins := range rsp.Data {
		c.insCache[ins.Symbol] = ins
	}

	snap := c.snapshot(0)
	result, _ := json.Marshal(snap.GetData())

	log.Info("Instrument partial: ", string(result))

	return changed
}

func (c *InstrumentCache) handleUpdate(data *ngerest.Instrument) (bool, error) {
	changed := false

	ins, exist := c.insCache[data.Symbol]

	if !exist {
		return false, errors.New("Partial data missing for " + data.Symbol)
	}

	if data.IndicativeSettlePrice > 0 && data.MarkPrice > 0 {
		c.applyInsPrice(data.IndicativeSettlePrice, data.MarkPrice)

		if ins.IndicativeSettlePrice != data.IndicativeSettlePrice {
			ins.IndicativeSettlePrice = data.IndicativeSettlePrice

			changed = true
		}

		if ins.MarkPrice != data.MarkPrice {
			ins.MarkPrice = data.MarkPrice

			changed = true
		}

		ins.FairBasis = data.FairBasis
		ins.FairPrice = data.FairPrice
		ins.PrevMarkPrice24H = data.PrevMarkPrice24H
	}

	if data.BidPrice > 0 {
		if ins.BidPrice != data.BidPrice {
			ins.BidPrice = data.BidPrice

			changed = true
		}

		if ins.AskPrice != data.AskPrice {
			ins.AskPrice = data.AskPrice

			changed = true
		}
	}

	if data.LastPrice > 0 {
		ins.LastPrice = data.LastPrice
		ins.PrevPrice24h = data.PrevPrice24h
		ins.LastChangePcnt = data.LastChangePcnt
		ins.LastTickDirection = data.LastTickDirection

		ins.Volume = data.Volume
		ins.Volume24h = data.Volume24h
		ins.TotalVolume = data.TotalVolume
		ins.PrevTotalVolume = data.PrevTotalVolume

		ins.Turnover = data.Turnover
		ins.Turnover24h = data.Turnover24h
		ins.TotalTurnover = data.TotalTurnover
		ins.PrevTotalTurnover = data.PrevTotalTurnover

		changed = true
	}

	if data.FundingRate > 0 {
		ins.FundingRate = data.FundingRate
		ins.FundingTimestamp = data.FundingTimestamp

		changed = true
	}

	return changed, nil
}

func (c *InstrumentCache) applyData(rsp *models.InstrumentResponse) bool {
	var (
		changed = false
		err     error
	)

	switch rsp.Action {
	case models.PartialAction:
		return c.handlePartial(rsp)
	case models.UpdateAction:
		for _, data := range rsp.Data {
			if changed, err = c.handleUpdate(data); err != nil {
				log.Error(err)
			}
		}
	default:
		log.Error("Invalid action for instrument cache: ", rsp.Action)
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
