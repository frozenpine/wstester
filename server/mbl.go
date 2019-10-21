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

	limitedChannel [3]Channel

	asks       sort.Float64Slice
	bids       sort.Float64Slice
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

	priceList := append([]float64{}, c.bids[buyLength-buyDepth:]...)
	priceList = append(priceList, c.asks[0:sellDepth]...)
	utils.ReverseFloat64Slice(priceList)

	dataList := make([]*ngerest.OrderBookL2, sellDepth+buyDepth)
	for idx, price := range priceList {
		dataList[idx] = c.orderCache[price]
	}

	snap.Data = dataList

	return snap
}

func (c *MBLCache) handleInput(in *CacheInput) {
	var rsp models.TableResponse

	if in.IsBreakPoint() {
		rsp = in.breakpointFunc()
	} else {
		mblNotify := kafka.MBLNotify{}

		if err := json.Unmarshal(in.msg, &mblNotify); err != nil {
			log.Println(err)
		} else {
			c.applyData(mblNotify.Content)
			rsp = mblNotify.Content
		}
	}

	if rsp == nil {
		return
	}

	if in.pubChannels != nil && len(in.pubChannels) > 0 {
		for _, ch := range in.pubChannels {
			ch.PublishData(rsp)
		}

		return
	}

	for chType, chGroup := range c.channelGroup {
		_ = chType

		for depth, ch := range chGroup {
			_ = depth

			ch.PublishData(rsp)
		}
	}
}

func (c *MBLCache) applyData(data *models.MBLResponse) {
	switch data.Action {
	case models.DeleteAction:
		for _, ord := range data.Data {
			if err := c.deleteOrder(ord); err != nil {
				log.Println(err)
			}
		}
	case models.InsertAction:
		for _, ord := range data.Data {
			if err := c.insertOrder(ord); err != nil {
				log.Println(err)
			}
		}
	case models.UpdateAction:
		for _, ord := range data.Data {
			if err := c.updateOrder(ord); err != nil {
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

func (c *MBLCache) deleteOrder(ord *ngerest.OrderBookL2) error {
	if _, exist := c.orderCache[ord.Price]; !exist {
		return fmt.Errorf("%s order[%f] delete on %s side not exist", ord.Symbol, ord.Price, ord.Side)
	}

	switch ord.Side {
	case "Buy":
		idx := c.bids.Search(ord.Price)
		c.bids = append(c.bids[0:idx], c.bids[idx+1:]...)
	case "Sell":
		idx := c.asks.Search(ord.Price)
		c.asks = append(c.asks[0:idx], c.asks[idx+1:]...)
	default:
		return errors.New("invalid order side: " + ord.Side)
	}

	delete(c.orderCache, ord.Price)

	return nil
}

func (c *MBLCache) insertOrder(ord *ngerest.OrderBookL2) error {
	if origin, exist := c.orderCache[ord.Price]; exist {
		return fmt.Errorf(
			"%s order[%f@%.0f] insert on %s side with already exist order[%f@%.0f %.0f]",
			origin.Symbol, origin.Price, origin.Size, ord.Side, origin.Price, origin.Size, origin.ID,
		)
	}

	var dst *sort.Float64Slice

	switch ord.Side {
	case "Buy":
		dst = &c.bids
	case "Sell":
		dst = &c.asks
	default:
		return errors.New("invalid order side: " + ord.Side)
	}

	if dst.Len() < 1 || ord.Price > (*dst)[dst.Len()-1] {
		*dst = append((*dst), ord.Price)
	} else if ord.Price < (*dst)[0] {
		*dst = append(sort.Float64Slice{ord.Price}, *dst...)
	} else {
		midIdx := dst.Len() / 2

		*dst = append(append((*dst)[0:midIdx], ord.Price), (*dst)[midIdx:]...)

		sort.Sort(*dst)
	}

	c.orderCache[ord.Price] = ord

	return nil
}

func (c *MBLCache) updateOrder(ord *ngerest.OrderBookL2) error {
	if origin, exist := c.orderCache[ord.Price]; exist {
		origin.Size = ord.Size
		origin.ID = ord.ID

		return nil
	}

	return fmt.Errorf("%s order[%f@%.0f] update on %s side not exist", ord.Symbol, ord.Price, ord.Size, ord.Side)
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
