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
	channel

	ready chan bool

	ctx context.Context

	maxLength int

	pipeline chan interface{}

	historyTrade []ngerest.Trade
}

// RecentTrade get recent trade data
func (c *TradeCache) RecentTrade() []ngerest.Trade {
	ch := make(chan []ngerest.Trade, 1)

	c.pipeline <- func() []ngerest.Trade {
		snap := c.snapshot()

		ch <- snap

		return snap
	}

	return <-ch
}

// Start start cache backgroud goroutine
func (c *TradeCache) Start() (err error) {
	// client, err := sarama.NewConsumerGroup()

	go func() {
		for obj := range c.pipeline {
			switch obj.(type) {
			case func() []ngerest.Trade:
				breakPointFunc := obj.(func() []ngerest.Trade)

				partial := models.NewTradePartial()
				partial.Data = breakPointFunc()

				c.PublishData(&Message{
					IsSnapshot: true,
					Data:       partial,
				})
			case *sarama.ConsumerMessage:
				msg := obj.(*sarama.ConsumerMessage)

				c.PublishData(&Message{
					IsSnapshot: false,
					Data:       c.applyMessage(msg),
				})
			default:
				log.Println("invalid pipeline object:", obj)
			}
		}
	}()

	err = c.channel.Start()

	return err

	// c.consumer.ready = make(chan bool, 0)

	// for {
	// 	if err := client.Consume(c.ctx); err != nil {
	// 		log.Panicln("Error from consumer:", err)
	// 	}

	// 	if c.ctx.Err() != nil {
	// 		return
	// 	}
	// }
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
		c.pipeline <- message
	}

	return nil
}

func (c *TradeCache) snapshot() []ngerest.Trade {
	hisLen := len(c.historyTrade)

	idx := utils.MinInt(c.maxLength, hisLen)

	return c.historyTrade[hisLen-idx:]
}

func (c *TradeCache) parseData(msg *sarama.ConsumerMessage) *models.TradeResponse {
	ob := models.TradeResponse{}

	log.Println(string(msg.Value))
	json.Unmarshal(msg.Value, &ob)

	return &ob
}

func (c *TradeCache) applyMessage(msg *sarama.ConsumerMessage) *models.TradeResponse {
	ob := c.parseData(msg)

	c.historyTrade = append(c.historyTrade, ob.Data...)

	if hisLen := len(c.historyTrade); hisLen > c.maxLength*3 {
		trimLen := int(float64(c.maxLength) * 1.5)

		c.historyTrade = c.historyTrade[hisLen-trimLen:]
	}

	return ob
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
