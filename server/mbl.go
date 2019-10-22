package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"sort"

	"github.com/frozenpine/ngerest"
	"github.com/frozenpine/wstester/kafka"
	"github.com/frozenpine/wstester/models"
	"github.com/frozenpine/wstester/utils"
)

// MBLCache retrive & store mbl data
type MBLCache struct {
	tableCache

	asks       sort.Float64Slice // in DESC order
	bids       sort.Float64Slice // in ASC order
	orderCache map[float64]*ngerest.OrderBookL2
}

func (c *MBLCache) snapshot(depth int) models.TableResponse {
	if depth < 1 {
		depth = math.MaxInt64
	}

	snap := models.NewMBLPartial()

	sellLength := c.asks.Len()
	buyLength := c.bids.Len()
	sellDepth := utils.MinInt(sellLength, depth)
	buyDepth := utils.MinInt(buyLength, depth)

	priceList := make([]float64, sellDepth+buyDepth)

	copy(priceList, c.asks[sellLength-sellDepth:])
	copy(priceList[sellDepth:], c.bids[buyLength-buyDepth:])

	utils.ReverseFloat64Slice(priceList[sellDepth:])

	dataList := make([]*ngerest.OrderBookL2, sellDepth+buyDepth)
	for idx, price := range priceList {
		dataList[idx] = c.orderCache[price]
	}

	snap.Data = dataList

	return snap
}

func (c *MBLCache) handleInput(in *CacheInput) {
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

	mblNotify := kafka.MBLNotify{}

	if err := json.Unmarshal(in.msg, &mblNotify); err != nil {
		log.Println(err)
	} else {
		c.applyData(mblNotify.Content)
	}

	// FIXME: 不用深度级别的notify分发
	for depth, ch := range c.channelGroup[Realtime] {
		_ = depth

		ch.PublishData(mblNotify.Content)
	}
}

func (c *MBLCache) applyData(data *models.MBLResponse) {
	switch data.Action {
	case models.DeleteAction:
		for _, ord := range data.Data {
			if _, err := c.deleteOrder(ord); err != nil {
				log.Println(err)
			}
		}
	case models.InsertAction:
		for _, ord := range data.Data {
			if _, err := c.insertOrder(ord); err != nil {
				log.Println(err)
			}
		}
	case models.UpdateAction:
		for _, ord := range data.Data {
			if _, err := c.updateOrder(ord); err != nil {
				log.Println(err)
			}
		}
	case models.PartialAction:
		c.initCache()

		for _, ord := range data.Data {
			c.insertOrder(ord)
		}
	default:
		log.Panicln("Invalid action:", data.Action)
	}
}

func (c *MBLCache) initCache() {
	c.orderCache = make(map[float64]*ngerest.OrderBookL2)
	c.asks = sort.Float64Slice{}
	c.bids = sort.Float64Slice{}
}

func (c *MBLCache) deleteOrder(ord *ngerest.OrderBookL2) (int, error) {
	if _, exist := c.orderCache[ord.Price]; !exist {
		return 0, fmt.Errorf("%s order[%f] delete on %s side not exist", ord.Symbol, ord.Price, ord.Side)
	}

	var (
		idx int
		dst *sort.Float64Slice
	)

	switch ord.Side {
	case "Buy":
		dst = &c.bids
	case "Sell":
		dst = &c.asks
	default:
		return 0, errors.New("invalid order side: " + ord.Side)
	}

	idx = dst.Search(ord.Price)
	*dst = append((*dst)[0:idx], (*dst)[idx+1:]...)

	delete(c.orderCache, ord.Price)

	return idx + 1, nil
}

func (c *MBLCache) insertOrder(ord *ngerest.OrderBookL2) (int, error) {
	if origin, exist := c.orderCache[ord.Price]; exist {
		return 0, fmt.Errorf(
			"%s order[%f@%.0f] insert on %s side with already exist order[%f@%.0f %.0f]",
			origin.Symbol, origin.Price, origin.Size, ord.Side, origin.Price, origin.Size, origin.ID,
		)
	}

	var (
		idx    int
		sorted sort.Float64Slice
	)

	switch ord.Side {
	case "Buy":
		idx, sorted = utils.PriceSort(c.bids, ord.Price, false)
		if idx < 0 {
			return 0, errors.New("fail to insert price in bids")
		}
		c.bids = sorted
	case "Sell":
		idx, sorted = utils.PriceSort(c.asks, ord.Price, true)
		if idx < 0 {
			return 0, errors.New("fail to insert price in price list")
		}
		c.asks = sorted
	default:
		return 0, errors.New("invalid order side: " + ord.Side)
	}

	c.orderCache[ord.Price] = ord

	return idx + 1, nil
}

func (c *MBLCache) updateOrder(ord *ngerest.OrderBookL2) (int, error) {
	if origin, exist := c.orderCache[ord.Price]; exist {
		var idx int

		switch ord.Side {
		case "Buy":
			idx = c.bids.Search(ord.Price)
		case "Sell":
			idx = c.asks.Search(ord.Price)
		default:
			return 0, errors.New("invalid side: " + ord.Side)
		}

		origin.Size = ord.Size
		origin.ID = ord.ID

		return idx + 1, nil
	}

	return 0, fmt.Errorf("%s order[%f@%.0f] update on %s side not exist", ord.Symbol, ord.Price, ord.Size, ord.Side)
}

// NewMBLCache make a new MBL cache.
func NewMBLCache(ctx context.Context) *MBLCache {
	mbl := MBLCache{}
	mbl.ctx = ctx
	mbl.handleInputFn = mbl.handleInput
	mbl.snapshotFn = mbl.snapshot

	mbl.initCache()

	if err := mbl.Start(); err != nil {
		log.Panicln(err)
	}

	return &mbl
}
