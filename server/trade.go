package server

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Shopify/sarama"
	"github.com/frozenpine/ngerest"
	"github.com/frozenpine/wstester/models"
	"github.com/frozenpine/wstester/utils"
)

var (
	defaultTradeLen int = 200
)

// TradeCache retrive & store trade data
type TradeCache struct {
	tableCache

	historyTrade []*ngerest.Trade
}

func (c *TradeCache) snapshot() []*ngerest.Trade {
	hisLen := len(c.historyTrade)

	idx := utils.MinInt(c.maxLength, hisLen)

	return c.historyTrade[hisLen-idx:]
}

func (c *TradeCache) parseData(msg *sarama.ConsumerMessage) *models.TradeResponse {
	rsp := models.TradeResponse{}

	parsed := make(map[string]interface{})

	json.Unmarshal(msg.Value, &parsed)

	return &rsp
}

func (c *TradeCache) applyData(data *models.TradeResponse) {
	c.historyTrade = append(c.historyTrade, data.Data...)

	if hisLen := len(c.historyTrade); hisLen > c.maxLength*3 {
		trimLen := int(float64(c.maxLength) * 1.5)

		c.historyTrade = c.historyTrade[hisLen-trimLen:]
	}
}

// NewTradeCache make a new trade cache.
func NewTradeCache(ctx context.Context) *TradeCache {
	td := TradeCache{}

	if err := td.Start(ctx); err != nil {
		log.Panicln(err)
	}

	return &td
}
