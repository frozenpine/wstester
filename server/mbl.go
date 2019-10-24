package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"github.com/frozenpine/ngerest"
	"github.com/frozenpine/wstester/client"
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

// BestBuy best bid price
func (c *MBLCache) BestBuy() float64 {
	if c.bids.Len() > 0 {
		return c.bids[c.bids.Len()-1]
	}

	return 0
}

// BestSell best ask price
func (c *MBLCache) BestSell() float64 {
	if c.asks.Len() > 0 {
		return c.asks[c.asks.Len()-1]
	}

	return 0
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
		return
	}

	if mblNotify.Content == nil {
		log.Println("MBL notify content is empty:", string(in.msg))
		return
	}

	depth, err := c.applyData(mblNotify.Content)

	if err != nil {
		log.Printf("invalid apply depth[%d] return: %s", depth, string(in.msg))
		return
	}

	// log.Println("mbl:", c.BestBuy(), c.BestSell())

	// apply an partial
	if depth == 0 {
		return
	}

	for lvl, ch := range c.channelGroup[Realtime] {
		if lvl > 0 && depth > lvl {
			continue
		}

		ch.PublishData(mblNotify.Content)
	}
}

func (c *MBLCache) applyData(data *models.MBLResponse) (int, error) {
	var (
		depth int
		err   error
		ord   *ngerest.OrderBookL2
	)

	switch data.Action {
	case models.DeleteAction:
		for _, ord = range data.Data {
			if depth, err = c.deleteOrder(ord); err != nil {
				log.Println(err)
			}
		}
	case models.InsertAction:
		for _, ord = range data.Data {
			if depth, err = c.insertOrder(ord); err != nil {
				log.Println(err)
			}
		}
	case models.UpdateAction:
		for _, ord = range data.Data {
			if depth, err = c.updateOrder(ord); err != nil {
				log.Println(err)
			}
		}
	case models.PartialAction:
		c.partial(data.Data)
		return 0, nil
	default:
		log.Panicln("Invalid action:", data.Action)
	}

	return depth, err
}

func (c *MBLCache) initCache() {
	c.orderCache = make(map[float64]*ngerest.OrderBookL2)
	c.asks = sort.Float64Slice{}
	c.bids = sort.Float64Slice{}
}

func (c *MBLCache) partial(data []*ngerest.OrderBookL2) {
	c.initCache()

	for _, mbl := range data {
		switch mbl.Side {
		case "Buy":
			c.bids = append(c.bids, mbl.Price)
		case "Sell":
			c.asks = append(c.asks, mbl.Price)
		default:
			log.Println("invalid mbl side:", mbl.Side)
			continue
		}

		c.orderCache[mbl.Price] = mbl
	}

	utils.ReverseFloat64Slice(c.bids)

	snap := c.snapshot(0).GetData()
	result, _ := json.Marshal(snap)

	log.Println("partial:", string(result))
}

func (c *MBLCache) deleteOrder(ord *ngerest.OrderBookL2) (int, error) {
	if _, exist := c.orderCache[ord.Price]; !exist {
		return 0, fmt.Errorf("%s order[%f] delete on %s side not exist", ord.Symbol, ord.Price, ord.Side)
	}

	var (
		idx int = -1
		err error
	)

	switch ord.Side {
	case "Buy":
		idx, c.bids = utils.PriceRemove(c.bids, ord.Price, false)
	case "Sell":
		idx, c.asks = utils.PriceRemove(c.asks, ord.Price, true)
	default:
		err = errors.New("invalid order side: " + ord.Side)
	}

	if idx < 0 {
		err = fmt.Errorf("price %f not found on delete %s", ord.Price, ord.Side)
	}

	delete(c.orderCache, ord.Price)

	return idx, err
}

func (c *MBLCache) insertOrder(ord *ngerest.OrderBookL2) (int, error) {
	if origin, exist := c.orderCache[ord.Price]; exist {
		return 0, fmt.Errorf(
			"%s order[%f@%.0f] insert on %s side with already exist order[%f@%.0f %.0f]",
			origin.Symbol, origin.Price, origin.Size, ord.Side, origin.Price, origin.Size, origin.ID,
		)
	}

	var (
		idx int = -1
		err error
	)

	switch ord.Side {
	case "Buy":
		idx, c.bids = utils.PriceAdd(c.bids, ord.Price, false)
	case "Sell":
		idx, c.asks = utils.PriceAdd(c.asks, ord.Price, true)
	default:
		err = errors.New("invalid order side: " + ord.Side)
	}

	c.orderCache[ord.Price] = ord

	return idx, err
}

func (c *MBLCache) updateOrder(ord *ngerest.OrderBookL2) (int, error) {
	var (
		idx int = -1
		err error
	)

	if origin, exist := c.orderCache[ord.Price]; exist {
		switch ord.Side {
		case "Buy":
			idx = utils.PriceSearch(c.bids, ord.Price, false)
		case "Sell":
			idx = utils.PriceSearch(c.asks, ord.Price, true)
		default:
			err = errors.New("invalid order side: " + ord.Side)
		}

		if idx < 0 {
			err = fmt.Errorf("price %f not found on %s", ord.Price, ord.Side)
		} else {
			origin.Size = ord.Size
			origin.ID = ord.ID
		}
	} else {
		err = fmt.Errorf("%s order[%f@%.0f] update on %s side not exist", ord.Symbol, ord.Price, ord.Size, ord.Side)
	}

	return idx, err
}

func mockMBL(cache Cache) {
	cfg := client.NewConfig()
	ins := client.NewClient(cfg)
	ins.Subscribe("orderBookL2")

	for {
		ctx, cancelFn := context.WithCancel(context.Background())

		ins.Connect(ctx)

		go func() {
			mblChan := ins.GetResponse("orderBookL2")

			for {
				select {
				case mbl, ok := <-mblChan:
					if !ok {
						cancelFn()
						return
					}

					if mbl == nil {
						continue
					}

					notify := kafka.MBLNotify{}
					notify.Type = "orderBookL2"
					notify.Content = mbl.(*models.MBLResponse)

					result, _ := json.Marshal(notify)

					cache.Append(NewCacheInput(result))
				}
			}
		}()

		<-ins.Closed()

		<-time.After(time.Second * 5)
	}
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

	// mbl.channelGroup[Realtime][25] = &rspChannel{ctx: ctx}
	// if err := mbl.channelGroup[Realtime][25].Start(); err != nil {
	// 	log.Panicln(err)
	// }

	return &mbl
}
