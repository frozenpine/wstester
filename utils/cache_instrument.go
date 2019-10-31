package utils

import (
	"github.com/frozenpine/ngerest"
	"github.com/frozenpine/wstester/models"
)

// InstrumentCache retrive & store instrument data
type InstrumentCache struct {
	tableCache

	instrument *ngerest.Instrument
}

func (ins *InstrumentCache) snapshot(depth int) models.TableResponse {
	rsp := models.NewInstrumentPartial()

	rsp.Data = append(rsp.Data, ins.instrument)

	return rsp
}

func (ins *InstrumentCache) handleInput(in *CacheInput) {
	if in.IsBreakPoint() {
		rsp := in.breakpointFunc()
		if rsp == nil {
			return
		}

		if in.pubChannel != nil {
			in.pubChannel.PublishDataToDestination(rsp, in.dstIdx)
		}

		return
	}
}

// NewInstrumentCache make a new instrument cache.
func NewInstrumentCache() *InstrumentCache {
	ins := InstrumentCache{}

	return &ins
}
