package server

import (
	"sync"

	"github.com/frozenpine/ngerest"
	"github.com/frozenpine/wstester/utils"
)

var (
	defaultTradeLen int = 200
)

// TradeCache retrive & store trade data
type TradeCache struct {
	channel

	maxLength int

	historyTrade []ngerest.OrderBookL2
	snapLock     sync.Mutex
}

// RecentTrade get recent trade data
func (c *TradeCache) RecentTrade() []ngerest.OrderBookL2 {
	c.snapLock.Lock()
	defer func() {
		c.snapLock.Unlock()
	}()

	hisLen := len(c.historyTrade)

	idx := utils.MinInt(c.maxLength, hisLen)

	return c.historyTrade[hisLen-idx:]
}

// NewTradeCache make a new trade cache.
func NewTradeCache() *TradeCache {
	td := TradeCache{}

	return &td
}
