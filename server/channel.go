package server

import (
	"fmt"
	"log"
	"sync"
)

// Channel message channel
type Channel interface {
	// Start initialize channel and start a dispatch goroutine
	Start() error
	// Close close channel
	Close() error
	// Connect connect child channel, child channel will get dispatched data from current channel
	Connect(Channel) error

	// PublishData publish data to current channel
	PublishData(interface{}) error
	// RetriveData to get an chan to retrive data in current channel
	RetriveData() <-chan interface{}
}

type channel struct {
	source       chan interface{}
	destinations []chan interface{}
	subChannels  []Channel
	tmpDest      sync.Pool
	tmpSubChan   sync.Pool

	dispatchLock sync.Mutex

	isStarted bool
	startOnce sync.Once

	isClosed  bool
	closeOnce sync.Once
}

func (c *channel) PublishData(data interface{}) error {
	if c.isClosed {
		return fmt.Errorf("channel is already closed")
	}

	if !c.isStarted {
		c.Start()
	}

	c.source <- data

	return nil
}

func (c *channel) RetriveData() <-chan interface{} {
	if c.isClosed {
		ch := make(chan interface{}, 0)
		close(ch)

		return ch
	}

	if !c.isStarted {
		c.Start()
	}

	ch := make(chan interface{})

	c.dispatchLock.Lock()
	c.destinations = append(c.destinations, ch)
	c.dispatchLock.Unlock()

	return ch
}

func (c *channel) Connect(child Channel) error {
	if c.isClosed {
		return fmt.Errorf("channel is already closed")
	}

	if !c.isStarted {
		c.Start()
	}

	c.dispatchLock.Lock()
	c.subChannels = append(c.subChannels, child)
	c.dispatchLock.Unlock()

	child.Start()

	return nil
}

func (c *channel) Start() error {
	if c.isStarted {
		return fmt.Errorf("channel is already started")
	}

	if c.isClosed {
		return fmt.Errorf("channel is already closed")
	}

	c.startOnce.Do(func() {
		if c.source == nil {
			c.source = make(chan interface{})
		}

		c.tmpDest.Put([]chan interface{}{})
		c.tmpSubChan.Put([]Channel{})

		c.isClosed = false

		go func() {
			defer func() {
				for _, ch := range c.destinations {
					close(ch)
				}

				for _, subChan := range c.subChannels {
					subChan.Close()
				}
			}()

			for data := range c.source {
				c.dispatchDistinations(data)
				c.dispatchSubChannels(data)
			}
		}()

		c.isStarted = true
	})

	return nil
}

func (c *channel) Close() error {
	if !c.isStarted {
		return fmt.Errorf("channel is not started")
	}

	if c.isClosed {
		return fmt.Errorf("channel is already closed")
	}

	c.closeOnce.Do(func() {
		c.isStarted = false

		close(c.source)

		c.isClosed = true
	})

	return nil
}

func (c *channel) dispatchDistinations(data interface{}) {
	normalDest := c.tmpDest.Get().([]chan interface{})

	c.dispatchLock.Lock()
	defer func() {
		c.dispatchLock.Unlock()
	}()

	for _, dest := range c.destinations {
		if dest == nil {
			continue
		}

		dest <- data
		normalDest = append(normalDest, dest)
	}

	if len(normalDest) < len(c.destinations) {
		c.destinations = make([]chan interface{}, len(normalDest))
		copy(c.destinations, normalDest)
	}
}

func (c *channel) dispatchSubChannels(data interface{}) {
	normalChan := c.tmpSubChan.Get().([]Channel)

	c.dispatchLock.Lock()
	defer func() {
		c.dispatchLock.Unlock()
	}()

	for _, subChan := range c.subChannels {
		if err := subChan.PublishData(data); err != nil {
			log.Println(err)
			continue
		}

		normalChan = append(normalChan, subChan)
	}

	if len(normalChan) < len(c.subChannels) {
		c.subChannels = make([]Channel, len(normalChan))
		copy(c.subChannels, normalChan)
	}
}
