package server

import (
	"context"
	"log"
	"sort"

	"github.com/frozenpine/ngerest"
	"github.com/frozenpine/wstester/models"
)

// MBLCache retrive & store mbl data
type MBLCache struct {
	tableCache

	orderCache map[float64]*ngerest.OrderBookL2
}

func (c *MBLCache) snapshot() models.TableResponse {
	snap := models.NewMBLPartial()

	dataList := make([]*ngerest.OrderBookL2, len(c.orderCache))

	priceList := []float64{}
	for price := range c.orderCache {
		priceList = append(priceList, price)
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(priceList)))

	for idx, price := range priceList {
		dataList[idx] = c.orderCache[price]
	}

	snap.Data = dataList

	return snap
}

func (c *MBLCache) handleInput(in *CacheInput) models.TableResponse {
	var rsp models.TableResponse

	if in.IsBreakPoint() {
		rsp = in.breakpointFunc()
	} else {
		mbl := models.MBLResponse{}

		// TODO: sub flow handle

		c.applyData(&mbl)
	}

	return rsp
}

func (c *MBLCache) applyData(data *models.MBLResponse) {

}

// NewMBLCache make a new MBL cache.
func NewMBLCache(ctx context.Context) *MBLCache {
	mbl := MBLCache{}

	mbl.pipeline = make(chan *CacheInput, 1000)
	mbl.destinations = make(map[Session]chan models.Response)
	mbl.ready = make(chan struct{})
	mbl.handleInputFn = mbl.handleInput
	mbl.snapshotFn = mbl.snapshot

	if err := mbl.Start(ctx); err != nil {
		log.Panicln(err)
	}

	return &mbl
}
