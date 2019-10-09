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

type tradeMessage struct {
	doNotPublish   bool
	breakPointFunc func() []*ngerest.Trade
	msg            *sarama.ConsumerMessage
}

func (msg *tradeMessage) IsBreakPoint() bool {
	return msg.breakPointFunc != nil
}

// TradeCache retrive & store trade data
type TradeCache struct {
	rspChannel

	pipeline     chan tradeMessage
	ready        chan bool
	ctx          context.Context
	maxLength    int
	historyTrade []*ngerest.Trade
}

// RecentTrade get recent trade data, goroutine safe
func (c *TradeCache) RecentTrade(publish bool) []*ngerest.Trade {
	ch := make(chan []*ngerest.Trade, 1)
	defer func() {
		close(ch)
	}()

	c.pipeline <- tradeMessage{
		doNotPublish: !publish,
		breakPointFunc: func() []*ngerest.Trade {
			snap := c.snapshot()

			ch <- snap

			return snap
		},
	}

	return <-ch
}

// Start start cache backgroud goroutine
func (c *TradeCache) Start() (err error) {
	go func() {
		var rsp *models.TradeResponse

		for obj := range c.pipeline {
			if obj.IsBreakPoint() {
				rsp = models.NewTradePartial()
				rsp.Data = obj.breakPointFunc()
			} else {
				rsp = c.parseData(obj.msg)
				c.applyData(rsp)
			}

			if !obj.doNotPublish {
				c.PublishData(rsp)
			}
		}
	}()

	err = c.rspChannel.Start()

	return err
}

// Setup setup for consumer
func (c *TradeCache) Setup(sarama.ConsumerGroupSession) error {
	close(c.ready)
	return nil
}

// Cleanup cleanup for consumer
func (c *TradeCache) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim consume message from claim
func (c *TradeCache) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		c.pipeline <- tradeMessage{
			msg: message,
		}
	}

	return nil
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
func NewTradeCache() *TradeCache {
	td := TradeCache{}

	if err := td.Start(); err != nil {
		log.Panicln(err)

		return nil
	}

	return &td
}
