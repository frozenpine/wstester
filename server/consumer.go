package server

import "github.com/Shopify/sarama"

type consumer struct {
	ready []chan bool
}

func (c *consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}
