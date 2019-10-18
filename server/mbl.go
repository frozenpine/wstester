package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"

	"github.com/frozenpine/ngerest"
	"github.com/frozenpine/wstester/kafka"
	"github.com/frozenpine/wstester/models"
	"github.com/frozenpine/wstester/utils"
)

// MBLCache retrive & store mbl data
type MBLCache struct {
	tableCache

	buyCount  int
	sellCount int

	orderCache map[float64]*ngerest.OrderBookL2
}

func (c *MBLCache) snapshot(depth int) models.TableResponse {
	snap := models.NewMBLPartial()

	dataList := make([]*ngerest.OrderBookL2, len(c.orderCache))

	priceList := []float64{}
	for price := range c.orderCache {
		priceList = append(priceList, price)
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(priceList)))

	sellStart := c.sellCount - utils.MinInt(c.sellCount, depth)
	buyEnd := c.sellCount + utils.MinInt(c.buyCount, depth)
	if depth <= 0 {
		sellStart = 0
		buyEnd = c.sellCount + c.buyCount
	}

	priceList = priceList[sellStart:buyEnd]

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
		mblNotify := kafka.MBLNotify{}

		if err := json.Unmarshal(in.msg, &mblNotify); err != nil {
			log.Println(err)
		} else {
			c.applyData(mblNotify.Content)
			rsp = mblNotify.Content
		}
	}

	return rsp
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
	c.sellCount = 0
	c.buyCount = 0
}

func (c *MBLCache) deleteOrder(ord *ngerest.OrderBookL2) error {
	if _, exist := c.orderCache[ord.Price]; !exist {
		return fmt.Errorf("%s order[%f] delete on %s side not exist", ord.Symbol, ord.Price, ord.Side)
	}

	switch ord.Side {
	case "Buy":
		c.buyCount--
	case "Sell":
		c.sellCount--
	default:
		return errors.New("invalid order side: " + ord.Side)
	}

	delete(c.orderCache, ord.Price)

	return nil
}

func (c *MBLCache) insertOrder(ord *ngerest.OrderBookL2) error {
	if orgin, exist := c.orderCache[ord.Price]; exist {
		return fmt.Errorf(
			"%s order[%f@%.0f] insert on %s side with already exist order[%f@%.0f %.0f]",
			orgin.Symbol, orgin.Price, orgin.Size, ord.Side, orgin.Price, orgin.Size, orgin.ID,
		)
	}

	switch ord.Side {
	case "Buy":
		c.buyCount++
	case "Sell":
		c.sellCount++
	default:
		return errors.New("invalid order side: " + ord.Side)
	}

	c.orderCache[ord.Price] = ord

	return nil
}

func (c *MBLCache) updateOrder(ord *ngerest.OrderBookL2) error {
	if orgin, exist := c.orderCache[ord.Price]; exist {
		orgin.Size = ord.Size
		orgin.ID = ord.ID

		return nil
	}

	return fmt.Errorf("%s order[%f@%.0f] update on %s side not exist", ord.Symbol, ord.Price, ord.Size, ord.Side)
}

// NewMBLCache make a new MBL cache.
func NewMBLCache(ctx context.Context) *MBLCache {
	mbl := MBLCache{}

	mbl.handleInputFn = mbl.handleInput
	mbl.snapshotFn = mbl.snapshot
	mbl.initCache()

	if err := mbl.Start(ctx); err != nil {
		log.Panicln(err)
	}

	return &mbl
}
