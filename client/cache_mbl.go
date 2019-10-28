package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"sort"

	"github.com/frozenpine/ngerest"
	"github.com/frozenpine/wstester/models"
	"github.com/frozenpine/wstester/utils"
)

// MBLCache retrive & store mbl data
type MBLCache struct {
	tableCache

	askPrices []float64 // in DESC order
	bidPrices []float64 // in ASC order
	askQuote  struct {
		bestPrice, lastPrice float64
		bestSize, lastSize   float32
	}
	bidQuote struct {
		bestPrice, lastPrice float64
		bestSize, lastSize   float32
	}
	l2Cache map[float64]*ngerest.OrderBookL2
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

	sellLength := len(c.askPrices)
	buyLength := len(c.bidPrices)
	sellDepth := utils.MinInt(sellLength, depth)
	buyDepth := utils.MinInt(buyLength, depth)

	priceList := make([]float64, sellDepth+buyDepth)

	copy(priceList, c.askPrices[sellLength-sellDepth:])
	copy(priceList[sellDepth:], c.bidPrices[buyLength-buyDepth:])

	utils.ReverseFloat64Slice(priceList[sellDepth:])

	dataList := make([]*ngerest.OrderBookL2, sellDepth+buyDepth)
	for idx, price := range priceList {
		dataList[idx] = c.l2Cache[price]
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
					ch <- rsp
				}
			}
		}

		return
	}

	limitRsp, err := c.applyData(in.msg.(*models.MBLResponse))

	if err != nil {
		log.Printf("apply data failed: %s, data: %s", err.Error(), in.msg.String())
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

	c.output <- in.msg
}

// GetDepth get side depth
func (c *MBLCache) GetDepth(side string) int {
	switch side {
	case "Buy":
		return len(c.bidPrices)
	case "Sell":
		return len(c.askPrices)
	default:
		log.Println("invalid side in GetDepth:", side)
		return -1
	}
}

// GetOrderOnDepth get mbl order on specified depth
func (c *MBLCache) GetOrderOnDepth(side string, depth int) *ngerest.OrderBookL2 {
	var depthPrice float64

	switch side {
	case "Buy":
		bidLength := len(c.bidPrices)

		if bidLength < 1 || depth > bidLength {
			return nil
		}

		depthPrice = c.bidPrices[bidLength-depth]
	case "Sell":
		askLength := len(c.askPrices)

		if askLength < 1 || depth > askLength {
			return nil
		}

		depthPrice = c.askPrices[askLength-depth]
	default:
		log.Println("invalid side in makeup:", side)
		return nil
	}

	if ord, exist := c.l2Cache[depthPrice]; exist {
		return ord
	}

	log.Printf("Order of depth price[%.1f] not found in cache: %v\n", depthPrice, c.l2Cache)
	return nil
}

func (c *MBLCache) applyData(data *models.MBLResponse) ([]int, error) {
	var (
		depth     int
		depthList []int
		err       error
	)

	// FIXME: 价格挡位丢失
	switch data.Action {
	case models.DeleteAction:
		for _, ord := range data.Data {
			if depth, err = c.deleteOrder(ord); err != nil {
				return nil, err
			}

			depthList = append(depthList, depth)
		}
	case models.InsertAction:
		for _, ord := range data.Data {
			if depth, err = c.insertOrder(ord); err != nil {
				return nil, err
			}

			depthList = append(depthList, depth)
		}
	case models.UpdateAction:
		for _, ord := range data.Data {
			if depth, err = c.updateOrder(ord); err != nil {
				return nil, err
			}

			depthList = append(depthList, depth)
		}
	case models.PartialAction:
		c.partial(data.Data)
		return nil, nil
	default:
		return nil, fmt.Errorf("Invalid action: %s", data.Action)
	}

	return depthList, err
}

func (c *MBLCache) initCache() {
	c.l2Cache = make(map[float64]*ngerest.OrderBookL2)
	c.askPrices = sort.Float64Slice{}
	c.bidPrices = sort.Float64Slice{}
}

func (c *MBLCache) partial(data []*ngerest.OrderBookL2) {
	c.initCache()

	for _, mbl := range data {
		switch mbl.Side {
		case "Buy":
			c.bidPrices = append(c.bidPrices, mbl.Price)
		case "Sell":
			c.askPrices = append(c.askPrices, mbl.Price)
		default:
			log.Println("invalid mbl side:", mbl.Side)
			continue
		}

		c.l2Cache[mbl.Price] = mbl
	}

	utils.ReverseFloat64Slice(c.bidPrices)

	c.bidQuote.bestPrice = c.bidPrices[len(c.bidPrices)-1]
	c.bidQuote.bestSize = c.l2Cache[c.bidQuote.bestPrice].Size
	c.askQuote.bestPrice = c.askPrices[len(c.askPrices)-1]
	c.askQuote.bestSize = c.l2Cache[c.askQuote.bestPrice].Size

	snap := c.snapshot(0).GetData()
	result, _ := json.Marshal(snap)

	log.Println("MBL partial:", string(result))
}

func (c *MBLCache) deleteOrder(ord *ngerest.OrderBookL2) (int, error) {
	if _, exist := c.l2Cache[ord.Price]; !exist {
		return 0, fmt.Errorf("%s order[%.1f] delete on %s side not exist", ord.Symbol, ord.Price, ord.Side)
	}

	var (
		idx   int = -1
		err   error
		depth int
	)

	switch ord.Side {
	case "Buy":
		originLen := len(c.bidPrices)
		idx, c.bidPrices = utils.PriceRemove(c.bidPrices, ord.Price, false)
		depth = originLen - idx

		if depth == 1 {
			c.bidQuote.lastPrice, c.bidQuote.bestPrice = c.bidQuote.bestPrice, c.bidPrices[len(c.bidPrices)-1]
			c.bidQuote.lastSize, c.bidQuote.bestSize = c.bidQuote.bestSize, c.l2Cache[c.bidQuote.bestPrice].Size
		}
	case "Sell":
		originLen := len(c.askPrices)
		idx, c.askPrices = utils.PriceRemove(c.askPrices, ord.Price, true)
		depth = originLen - idx

		if depth == 1 {
			c.askQuote.lastPrice, c.askQuote.bestPrice = c.askQuote.bestPrice, c.askPrices[len(c.askPrices)-1]
			c.askQuote.lastSize, c.askQuote.bestSize = c.askQuote.bestSize, c.l2Cache[c.askQuote.bestPrice].Size
		}
	default:
		err = errors.New("invalid order side: " + ord.Side)
	}

	if idx < 0 {
		err = fmt.Errorf("price %f not found on delete %s", ord.Price, ord.Side)
	}

	delete(c.l2Cache, ord.Price)

	return depth, err
}

func (c *MBLCache) insertOrder(ord *ngerest.OrderBookL2) (int, error) {
	if origin, exist := c.l2Cache[ord.Price]; exist {
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
		idx, c.bidPrices = utils.PriceAdd(c.bidPrices, ord.Price, false)
		newLength := len(c.bidPrices)
		depth = newLength - idx

		if depth == 1 {
			c.bidQuote.lastPrice, c.bidQuote.bestPrice = c.bidQuote.bestPrice, ord.Price
			c.bidQuote.lastSize, c.bidQuote.bestSize = c.bidQuote.bestSize, ord.Size
		}
	case "Sell":
		idx, c.askPrices = utils.PriceAdd(c.askPrices, ord.Price, true)
		newLength := len(c.askPrices)
		depth = newLength - idx

		if depth == 1 {
			c.askQuote.lastPrice, c.askQuote.bestPrice = c.askQuote.bestPrice, ord.Price
			c.askQuote.lastSize, c.askQuote.bestSize = c.askQuote.bestSize, ord.Size
		}
	default:
		err = errors.New("invalid order side: " + ord.Side)
	}

	c.l2Cache[ord.Price] = ord

	return depth, err
}

func (c *MBLCache) updateOrder(ord *ngerest.OrderBookL2) (int, error) {
	var (
		idx   int = -1
		err   error
		depth int
	)

	if origin, exist := c.l2Cache[ord.Price]; exist {
		switch ord.Side {
		case "Buy":
			length := len(c.bidPrices)
			idx = utils.PriceSearch(c.bidPrices, ord.Price, false)
			depth = length - idx

			if depth == 1 {
				c.bidQuote.lastSize, c.bidQuote.bestSize = c.bidQuote.bestSize, ord.Size
			}

		case "Sell":
			length := len(c.askPrices)
			idx = utils.PriceSearch(c.askPrices, ord.Price, true)
			depth = length - idx

			if depth == 1 {
				c.askQuote.lastSize, c.askQuote.bestSize = c.askQuote.bestSize, ord.Size
			}
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
