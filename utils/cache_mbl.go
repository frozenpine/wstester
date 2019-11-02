package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/frozenpine/ngerest"
	"github.com/frozenpine/wstester/models"
	"github.com/frozenpine/wstester/utils/log"
)

// MBLCache retrive & store mbl data
type MBLCache struct {
	tableCache

	historyCount int64

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
	var (
		sellLength, buyLength, sellDepth, buyDepth int
	)

	snap := models.NewMBLPartial()

	sellLength = len(c.askPrices)
	buyLength = len(c.bidPrices)

	if depth > 0 {
		sellDepth = MinInt(sellLength, depth)
		buyDepth = MinInt(buyLength, depth)

		snap.Table = fmt.Sprintf("%s_%d", snap.Table, depth)
	} else {
		sellDepth = sellLength
		buyDepth = buyLength
	}

	priceList := make([]float64, sellDepth+buyDepth)

	copy(priceList, c.askPrices[sellLength-sellDepth:])
	copy(priceList[sellDepth:], c.bidPrices[buyLength-buyDepth:])

	ReverseFloat64Slice(priceList[sellDepth:])

	dataList := make([]*ngerest.OrderBookL2, sellDepth+buyDepth)
	for idx, price := range priceList {
		dataList[idx] = c.l2Cache[price]
	}

	snap.Data = dataList

	return snap
}

func (c *MBLCache) dispatchRsp(mbl *models.MBLResponse, limitRsp map[int][2]*models.MBLResponse) {
	if c.IsQuoteChange() {
		log.Debugf("Best Buy: %.1f@%.0f, Best Sell: %.1f@%.0f",
			c.BestBidPrice(), c.BestBidSize(), c.BestAskPrice(), c.BestAskSize())
	}

	// apply an partial
	if limitRsp == nil {
		return
	}

	for depth, ch := range c.channelGroup[Realtime] {
		if depth == 0 {
			ch.PublishData(mbl)
			continue
		}

		if rspList, exist := limitRsp[depth]; exist {
			for _, rsp := range rspList {
				if rsp != nil && len(rsp.Data) > 0 {
					rsp.Table = fmt.Sprintf("%s_%d", rsp.Table, depth)
					ch.PublishData(rsp)
				}
			}
		}
	}
}

func (c *MBLCache) handleInput(input *CacheInput) {
	if c.handleBreakpoint(input) {
		return
	}

	if input.msg == nil {
		log.DPanic("MBL notify content is empty:", input.msg.String())
		return
	}

	if mbl, ok := input.msg.(*models.MBLResponse); ok {
		c.historyCount += int64(len(mbl.Data))

		limitRsp, err := c.applyData(mbl)

		if err != nil {
			log.DPanic("apply data failed: %s, data: %s", err.Error(), input.msg.String())
			return
		}

		c.dispatchRsp(mbl, limitRsp)
	} else {
		log.DPanic("Can not convert cache input to MBLResponse.", input.msg.String())
	}
}

// GetDepth get side depth
func (c *MBLCache) GetDepth(side string) int {
	switch side {
	case "Buy":
		return len(c.bidPrices)
	case "Sell":
		return len(c.askPrices)
	default:
		log.DPanic("invalid side in GetDepth:", side)
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
		log.DPanic("invalid side in makeup:", side)
		return nil
	}

	if ord, exist := c.l2Cache[depthPrice]; exist {
		return ord
	}

	log.Error("Order of depth price[%.1f] not found in cache: %v\n", depthPrice, c.l2Cache)
	return nil
}

func (c *MBLCache) applyData(data *models.MBLResponse) (map[int][2]*models.MBLResponse, error) {
	var (
		depth int
		err   error
	)

	limitRspMap := make(map[int][2]*models.MBLResponse)

	for depth := range c.channelGroup[Realtime] {
		if depth != 0 {
			limitRsp := models.MBLResponse{}
			limitRsp.Table = data.Table
			limitRsp.Action = data.Action

			var makeupRsp *models.MBLResponse
			switch data.Action {
			case "delete":
				makeupRsp = &models.MBLResponse{}
				makeupRsp.Action = "insert"
				makeupRsp.Table = data.Table
			case "insert":
				makeupRsp = &models.MBLResponse{}
				makeupRsp.Action = "delete"
				makeupRsp.Table = data.Table
			default:
				makeupRsp = nil
			}

			limitRspMap[depth] = [2]*models.MBLResponse{&limitRsp, makeupRsp}
		}
	}

	switch data.Action {
	case models.DeleteAction:
		for _, ord := range data.Data {
			if depth, err = c.handleDelete(ord); err != nil {
				return nil, err
			}

			for limit, rspList := range limitRspMap {
				if depth <= limit {
					rspList[0].Data = append(rspList[0].Data, ord)

					makeupOrd := c.GetOrderOnDepth(ord.Side, limit)
					if makeupOrd != nil {
						rspList[1].Data = append(rspList[1].Data, makeupOrd)
					}
				}
			}
		}
	case models.InsertAction:
		for _, ord := range data.Data {
			if depth, err = c.handleInsert(ord); err != nil {
				return nil, err
			}

			for limit, rspList := range limitRspMap {
				if depth <= limit {
					rspList[0].Data = append(rspList[0].Data, ord)

					makeupOrd := c.GetOrderOnDepth(ord.Side, limit+1)
					if makeupOrd != nil {
						rspList[1].Data = append(rspList[1].Data, makeupOrd)
					}
				}
			}
		}
	case models.UpdateAction:
		for _, ord := range data.Data {
			if depth, err = c.handleUpdate(ord); err != nil {
				return nil, err
			}

			for limit, rspList := range limitRspMap {
				if depth <= limit {
					rspList[0].Data = append(rspList[0].Data, ord)
				}
			}
		}
	case models.PartialAction:
		c.handlePartial(data.Data)
		return nil, nil
	default:
		return nil, fmt.Errorf("Invalid action: %s", data.Action)
	}

	_ = depth

	return limitRspMap, err
}

func (c *MBLCache) initCache() {
	c.l2Cache = make(map[float64]*ngerest.OrderBookL2)
	c.askPrices = []float64{}
	c.bidPrices = []float64{}
}

func (c *MBLCache) handlePartial(data []*ngerest.OrderBookL2) {
	if len(c.l2Cache) < 1 {
		c.initCache()

		for _, mbl := range data {
			switch mbl.Side {
			case "Buy":
				c.bidPrices = append(c.bidPrices, mbl.Price)
			case "Sell":
				c.askPrices = append(c.askPrices, mbl.Price)
			default:
				log.DPanic("invalid mbl side:", mbl.Side)
				continue
			}

			c.l2Cache[mbl.Price] = mbl
		}

		ReverseFloat64Slice(c.bidPrices)

		c.bidQuote.bestPrice = c.bidPrices[len(c.bidPrices)-1]
		c.bidQuote.bestSize = c.l2Cache[c.bidQuote.bestPrice].Size
		c.askQuote.bestPrice = c.askPrices[len(c.askPrices)-1]
		c.askQuote.bestSize = c.l2Cache[c.askQuote.bestPrice].Size

		snap := c.snapshot(0).GetData()
		result, _ := json.Marshal(snap)

		log.Debug("MBL partial:", string(result))

		return
	}

	newSellSet := NewFloat64Set(nil)
	newBuySet := NewFloat64Set(nil)

	for _, ord := range data {
		if ord.Side == "Sell" {
			newSellSet.Append(ord.Price)
		}

		if ord.Side == "Buy" {
			newBuySet.Append(ord.Price)
		}
	}

	originSellSet := NewFloat64Set(c.askPrices)
	originBuySet := NewFloat64Set(c.bidPrices)

	sellAddSet := newSellSet.Sub(originSellSet).(Float64Set)
	buyAddSet := newBuySet.Sub(originBuySet).(Float64Set)

	sellDelSet := originSellSet.Sub(newSellSet).(Float64Set)
	buyDelSet := originBuySet.Sub(newBuySet).(Float64Set)

	sellUpdSet := originSellSet.Join(newSellSet).(Float64Set)
	buyUpdSet := originBuySet.Join(newBuySet).(Float64Set)

	insertRsp := models.MBLResponse{}
	insertRsp.Table = "orderBookL2"
	insertRsp.Action = models.InsertAction

	deleteRsp := models.MBLResponse{}
	deleteRsp.Table = "orderBookL2"
	deleteRsp.Action = models.DeleteAction

	updateRsp := models.MBLResponse{}
	updateRsp.Table = "orderBookL2"
	updateRsp.Action = models.UpdateAction

	for _, price := range sellDelSet.Add(buyDelSet).(Float64Set).Values() {
		deleteRsp.Data = append(deleteRsp.Data, c.l2Cache[price])
	}

	if len(deleteRsp.Data) > 0 {
		limitRsp, err := c.applyData(&deleteRsp)
		if err != nil {
			log.DPanic("Merge new partial delete failed:", err)
		} else {
			c.dispatchRsp(&deleteRsp, limitRsp)
			log.Info("New parital delete merged:", deleteRsp.String())
		}
	}

	for _, ord := range data {
		if sellAddSet.Exist(ord.Price) || buyAddSet.Exist(ord.Price) {
			insertRsp.Data = append(insertRsp.Data, ord)
		}

		if (sellUpdSet.Exist(ord.Price) && c.l2Cache[ord.Price].Size != ord.Size) ||
			(buyUpdSet.Exist(ord.Price) && c.l2Cache[ord.Price].Size != ord.Size) {
			updateRsp.Data = append(updateRsp.Data, ord)
		}
	}

	if len(insertRsp.Data) > 0 {
		limitRsp, err := c.applyData(&insertRsp)
		if err != nil {
			log.DPanic("Merge new partial insert failed:", err)
		} else {
			c.dispatchRsp(&insertRsp, limitRsp)
			log.Info("New parital insert merged:", insertRsp.String())
		}
	}

	if len(updateRsp.Data) > 0 {
		limitRsp, err := c.applyData(&updateRsp)
		if err != nil {
			log.DPanic("Merge new partial update failed:", err)
		} else {
			c.dispatchRsp(&updateRsp, limitRsp)
			log.Info("New parital update merged:", updateRsp.String())
		}
	}
}

func (c *MBLCache) handleDelete(ord *ngerest.OrderBookL2) (int, error) {
	if origin, exist := c.l2Cache[ord.Price]; !exist {
		return 0, fmt.Errorf("%s order[%.1f] delete on %s side not exist", ord.Symbol, ord.Price, ord.Side)
	} else if ord.ID != origin.ID {
		log.DPanic("order id miss-match with cache:", ord.ID, origin.ID)
	}

	var (
		idx   int = -1
		err   error
		depth int
	)

	switch ord.Side {
	case "Buy":
		originLen := len(c.bidPrices)
		idx, c.bidPrices = UniPriceRemove(c.bidPrices, ord.Price, false)
		depth = originLen - idx

		if depth == 1 {
			c.bidQuote.lastPrice, c.bidQuote.bestPrice = c.bidQuote.bestPrice, c.bidPrices[len(c.bidPrices)-1]
			c.bidQuote.lastSize, c.bidQuote.bestSize = c.bidQuote.bestSize, c.l2Cache[c.bidQuote.bestPrice].Size
		}
	case "Sell":
		originLen := len(c.askPrices)
		idx, c.askPrices = UniPriceRemove(c.askPrices, ord.Price, true)
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

func (c *MBLCache) handleInsert(ord *ngerest.OrderBookL2) (int, error) {
	if origin, exist := c.l2Cache[ord.Price]; exist {
		return 0, fmt.Errorf(
			"%s order[%.1f@%.0f] insert on %s side with already exist order[%.1f@%.0f %d]",
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
		idx, c.bidPrices = UniPriceAdd(c.bidPrices, ord.Price, false)
		newLength := len(c.bidPrices)
		depth = newLength - idx

		if depth == 1 {
			c.bidQuote.lastPrice, c.bidQuote.bestPrice = c.bidQuote.bestPrice, ord.Price
			c.bidQuote.lastSize, c.bidQuote.bestSize = c.bidQuote.bestSize, ord.Size
		}
	case "Sell":
		idx, c.askPrices = UniPriceAdd(c.askPrices, ord.Price, true)
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

func (c *MBLCache) handleUpdate(ord *ngerest.OrderBookL2) (int, error) {
	var (
		idx   int = -1
		err   error
		depth int
	)

	if origin, exist := c.l2Cache[ord.Price]; exist {
		switch ord.Side {
		case "Buy":
			length := len(c.bidPrices)
			idx = UniPriceSearch(c.bidPrices, ord.Price, false)
			depth = length - idx

			if depth == 1 {
				c.bidQuote.lastSize, c.bidQuote.bestSize = c.bidQuote.bestSize, ord.Size
			}

		case "Sell":
			length := len(c.askPrices)
			idx = UniPriceSearch(c.askPrices, ord.Price, true)
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
		}
	} else {
		err = fmt.Errorf("%s order[%.1f@%.0f] update on %s side not exist", ord.Symbol, ord.Price, ord.Size, ord.Side)
	}

	return depth, err
}

// NewDepthChannel create an new depth channel in cache
func (c *MBLCache) NewDepthChannel(depth int) error {
	c.channelGroup[Realtime][depth] = &rspChannel{ctx: c.ctx, retriveLock: sync.Mutex{}, connectLock: sync.Mutex{}}

	if err := c.channelGroup[Realtime][25].Start(); err != nil {
		return err
	}

	return nil
}

// NewMBLCache make a new MBL cache.
func NewMBLCache(ctx context.Context, symbol string) Cache {
	if ctx == nil {
		ctx = context.Background()
	}

	mbl := MBLCache{}
	mbl.Symbol = symbol
	mbl.ctx = ctx
	mbl.handleInputFn = mbl.handleInput
	mbl.snapshotFn = mbl.snapshot
	mbl.pipeline = make(chan *CacheInput, 1000)
	mbl.ready = make(chan struct{})
	mbl.channelGroup[Realtime] = map[int]Channel{
		0: &rspChannel{ctx: ctx, retriveLock: sync.Mutex{}, connectLock: sync.Mutex{}},
	}

	if err := mbl.Start(); err != nil {
		log.Panic(err)
	}

	return &mbl
}
