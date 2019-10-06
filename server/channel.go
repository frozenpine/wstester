package server

import (
	"fmt"
	"sync"
)

// Channel message channel
type Channel interface {
	Start()
	Close()

	PublishData(interface{}) error
	SubscribeData() <-chan interface{}

	Connect(Channel) error
}

type channel struct {
	source       chan interface{}
	destinations []chan interface{}

	subChannels []Channel

	isClosed  bool
	closeOnce sync.Once
}

func (c *channel) PublishData(data interface{}) error {
	if c.isClosed {
		return fmt.Errorf("channel is already closed")
	}

	c.source <- data

	return nil
}

func (c *channel) SubscribeData() <-chan interface{} {
	ch := make(chan interface{})
	c.destinations = append(c.destinations, ch)

	return ch
}

func (c *channel) Connect(child Channel) {
	c.subChannels = append(c.subChannels, child)
}

func (c *channel) Start() {
	go func() {
		for data := range c.source {
			c.dispatchDistinations(data)
		}
	}()
}

func (c *channel) dispatchDistinations(data interface{}) {
	for _, dest := range c.destinations {
		dest <- data
	}
}
