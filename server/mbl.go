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

	asks     []float64 // in DESC order
	bids     []float64 // in ASC order
	askQuote struct {
		bestPrice, lastPrice float64
		bestSize, lastSize   float32
	}
	bidQuote struct {
		bestPrice, lastPrice float64
		bestSize, lastSize   float32
	}
	orderCache map[float64]*ngerest.OrderBookL2
}

// BestBidPrice best bid price
func (c *MBLCache) BestBidPrice() float64 {
	return c.bidQuote.bestPrice
}

// BestBidSize best bid size
func (c *MBLCache) BestBidSize() float32 {
	return c.bidQuote.bestSize
}

// BestAskPrice best ask price
func (c *MBLCache) BestAskPrice() float64 {
	return c.askQuote.bestPrice
}

// BestAskSize best ask size
func (c *MBLCache) BestAskSize() float32 {
	return c.askQuote.bestSize
}

// IsQuoteChange true if best quote changed
func (c *MBLCache) IsQuoteChange() bool {
	bidChanged := c.bidQuote.bestPrice != c.bidQuote.lastPrice || c.bidQuote.bestSize != c.bidQuote.lastSize
	askChanged := c.askQuote.bestPrice != c.askQuote.lastPrice || c.askQuote.bestSize != c.askQuote.lastSize

	if bidChanged {
		c.bidQuote.lastPrice = c.bidQuote.bestPrice
		c.bidQuote.lastSize = c.bidQuote.bestSize
	}

	if askChanged {
		c.askQuote.lastPrice = c.askQuote.bestPrice
		c.askQuote.lastSize = c.askQuote.bestSize
	}

	return bidChanged || askChanged
}

func (c *MBLCache) snapshot(depth int) models.TableResponse {
	if depth < 1 {
		depth = math.MaxInt64
	}

	snap := models.NewMBLPartial()

	sellLength := len(c.asks)
	buyLength := len(c.bids)
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

	if depth > 0 && depth != math.MaxInt64 {
		snap.Table = fmt.Sprintf("%s_%d", snap.Table, depth)
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

	limitRsp, err := c.applyData(mblNotify.Content)

	if err != nil {
		log.Printf("apply data failed: %s, data: %s", err.Error(), string(in.msg))
		return
	}

	if c.IsQuoteChange() {
		log.Printf("Best Buy: %.1f@%.0f, Best Sell: %.1f@%.0f\n",
			c.BestBidPrice(), c.BestBidSize(), c.BestAskPrice(), c.BestAskSize())
	}

	// apply an partial
	if limitRsp == nil {
		return
	}

	for lvl, ch := range c.channelGroup[Realtime] {
		if lvl == 0 {
			ch.PublishData(mblNotify.Content)
			continue
		}

		if rsp, exist := limitRsp[lvl]; exist && len(rsp.Data) > 0 {
			rsp.Table = fmt.Sprintf("%s_%d", rsp.Table, lvl)
			ch.PublishData(rsp)
		}
	}
}

func (c *MBLCache) applyData(data *models.MBLResponse) (map[int]*models.MBLResponse, error) {
	var (
		depth int
		err   error
	)

	limitRsp := make(map[int]*models.MBLResponse)

	for depth := range c.channelGroup[Realtime] {
		if depth != 0 {
			rsp := models.MBLResponse{}
			rsp.Table = data.Table
			rsp.Action = data.Action

			limitRsp[depth] = &rsp
		}
	}

	switch data.Action {
	case models.DeleteAction:
		for _, ord := range data.Data {
			if depth, err = c.deleteOrder(ord); err != nil {
				return nil, err
			}

			for limit, rsp := range limitRsp {
				if depth <= limit {
					rsp.Data = append(rsp.Data, ord)
				}
			}
		}
	case models.InsertAction:
		for _, ord := range data.Data {
			if depth, err = c.insertOrder(ord); err != nil {
				return nil, err
			}

			for limit, rsp := range limitRsp {
				if depth <= limit {
					rsp.Data = append(rsp.Data, ord)
				}
			}
		}
	case models.UpdateAction:
		for _, ord := range data.Data {
			if depth, err = c.updateOrder(ord); err != nil {
				return nil, err
			}

			for limit, rsp := range limitRsp {
				if depth <= limit {
					rsp.Data = append(rsp.Data, ord)
				}
			}
		}
	case models.PartialAction:
		c.partial(data.Data)
		return nil, nil
	default:
		return nil, fmt.Errorf("Invalid action: %s", data.Action)
	}

	return limitRsp, err
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

	c.bidQuote.bestPrice = c.bids[len(c.bids)-1]
	c.bidQuote.bestSize = c.orderCache[c.bidQuote.bestPrice].Size
	c.askQuote.bestPrice = c.asks[len(c.asks)-1]
	c.askQuote.bestSize = c.orderCache[c.askQuote.bestPrice].Size

	snap := c.snapshot(0).GetData()
	result, _ := json.Marshal(snap)

	log.Println("MBL partial:", string(result))
}

func (c *MBLCache) deleteOrder(ord *ngerest.OrderBookL2) (int, error) {
	if _, exist := c.orderCache[ord.Price]; !exist {
		return 0, fmt.Errorf("%s order[%.1f] delete on %s side not exist", ord.Symbol, ord.Price, ord.Side)
	}

	var (
		idx   int = -1
		err   error
		depth int
	)

	switch ord.Side {
	case "Buy":
		originLen := len(c.bids)
		idx, c.bids = utils.PriceRemove(c.bids, ord.Price, false)
		if idx == originLen-1 {
			c.bidQuote.lastPrice, c.bidQuote.bestPrice = c.bidQuote.bestPrice, c.bids[len(c.bids)-1]
			c.bidQuote.lastSize, c.bidQuote.bestSize = c.bidQuote.bestSize, c.orderCache[c.bidQuote.bestPrice].Size
		}
		depth = originLen - idx
	case "Sell":
		originLen := len(c.asks)
		idx, c.asks = utils.PriceRemove(c.asks, ord.Price, true)
		if idx == originLen-1 {
			c.askQuote.lastPrice, c.askQuote.bestPrice = c.askQuote.bestPrice, c.asks[len(c.asks)-1]
			c.askQuote.lastSize, c.askQuote.bestSize = c.askQuote.bestSize, c.orderCache[c.askQuote.bestPrice].Size
		}
		depth = originLen - idx
	default:
		err = errors.New("invalid order side: " + ord.Side)
	}

	if idx < 0 {
		err = fmt.Errorf("price %f not found on delete %s", ord.Price, ord.Side)
	}

	delete(c.orderCache, ord.Price)

	return depth, err
}

func (c *MBLCache) insertOrder(ord *ngerest.OrderBookL2) (int, error) {
	if origin, exist := c.orderCache[ord.Price]; exist {
		return 0, fmt.Errorf(
			"%s order[%.1f@%.0f] insert on %s side with already exist order[%.1f@%.0f %.0f]",
			origin.Symbol, origin.Price, origin.Size, ord.Side, origin.Price, origin.Size, origin.ID,
		)
	}

	var (
		idx   int = -1
		err   error
		depth int
	)

	switch ord.Side {
	case "Buy":
		idx, c.bids = utils.PriceAdd(c.bids, ord.Price, false)
		newLength := len(c.bids)
		if idx == newLength-1 {
			c.bidQuote.lastPrice, c.bidQuote.bestPrice = c.bidQuote.bestPrice, ord.Price
			c.bidQuote.lastSize, c.bidQuote.bestSize = c.bidQuote.bestSize, ord.Size
		}
		depth = newLength - idx
	case "Sell":
		idx, c.asks = utils.PriceAdd(c.asks, ord.Price, true)
		newLength := len(c.asks)
		if idx == newLength-1 {
			c.askQuote.lastPrice, c.askQuote.bestPrice = c.askQuote.bestPrice, ord.Price
			c.askQuote.lastSize, c.askQuote.bestSize = c.askQuote.bestSize, ord.Size
		}
		depth = newLength - idx
	default:
		err = errors.New("invalid order side: " + ord.Side)
	}

	c.orderCache[ord.Price] = ord

	return depth, err
}

func (c *MBLCache) updateOrder(ord *ngerest.OrderBookL2) (int, error) {
	var (
		idx   int = -1
		err   error
		depth int
	)

	if origin, exist := c.orderCache[ord.Price]; exist {
		switch ord.Side {
		case "Buy":
			length := len(c.bids)
			idx = utils.PriceSearch(c.bids, ord.Price, false)
			if idx == length-1 {
				c.bidQuote.lastSize, c.bidQuote.bestSize = c.bidQuote.bestSize, ord.Size
			}
			depth = length - idx
		case "Sell":
			length := len(c.asks)
			idx = utils.PriceSearch(c.asks, ord.Price, true)
			if idx == length-1 {
				c.askQuote.lastSize, c.askQuote.bestSize = c.askQuote.bestSize, ord.Size
			}
			depth = length - idx
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
		err = fmt.Errorf("%s order[%.1f@%.0f] update on %s side not exist", ord.Symbol, ord.Price, ord.Size, ord.Side)
	}

	return depth, err
}

func mockMBL(cache Cache) {
	for {
		cfg := client.NewConfig()
		ins := client.NewClient(cfg)
		ins.Subscribe("orderBookL2")

		ctx, cancelFn := context.WithCancel(context.Background())

		ins.Connect(ctx)

		go func() {
			mblChan := ins.GetResponse("orderBookL2")

			for {
				select {
				case <-ctx.Done():
					return
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

		cancelFn()

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

	mbl.channelGroup[Realtime][25] = &rspChannel{ctx: ctx}
	if err := mbl.channelGroup[Realtime][25].Start(); err != nil {
		log.Panicln(err)
	}

	return &mbl
}
