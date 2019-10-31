package utils

import (
	"context"
	"log"
	"sync"

	"github.com/frozenpine/ngerest"
	"github.com/frozenpine/wstester/models"
)

// InstrumentCache retrive & store instrument data
type InstrumentCache struct {
	tableCache

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
		c.applyData(ins)

		c.channelGroup[Realtime][0].PublishData(ins)
	} else {
		log.Println("Can not convert cache input to InstrumentResponse:", input.msg.String())
	}
}

func (c *InstrumentCache) applyData(ins *models.InstrumentResponse) {
	switch ins.Action {
	case models.PartialAction:
		c.instrument = ins.Data[0]
	case models.UpdateAction:
		// TODO: UPDATE action handle
	default:
		log.Println("Invalid action for instrument cache:", ins.Action)
	}
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
		0: &rspChannel{ctx: ctx, retriveLock: sync.Mutex{}, connectLock: sync.Mutex{}},
	}

	if err := ins.Start(); err != nil {
		log.Panicln(err)
	}

	return &ins
}
