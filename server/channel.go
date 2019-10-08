package server

import (
	"fmt"
	"log"
	"sync"

	"github.com/frozenpine/wstester/models"
	"github.com/frozenpine/wstester/utils"
)

// Message data wrapper for identify weather data is a snapshot
type Message struct {
	IsSnapshot bool
	Data       interface{}
}

// Channel message channel
type Channel interface {
	// Start initialize channel and start a dispatch goroutine
	Start() error
	// Close close channel
	Close() error
	// Connect connect child channel, child channel will get dispatched data from current channel
	Connect(Channel) error

	// PublishData publish data to current channel
	PublishData(models.Response) error
	// RetriveData to get an chan to retrive data in current channel
	RetriveData() <-chan models.Response
}

type rspChannel struct {
	source chan models.Response

	destinations    []chan models.Response
	newDestinations []chan models.Response
	subChannels     []Channel
	newSubChannels  []Channel

	retriveLock sync.Mutex
	connectLock sync.Mutex

	isStarted bool
	startOnce sync.Once

	isClosed  bool
	closeOnce sync.Once
}

func (c *rspChannel) PublishData(data models.Response) error {
	if c.isClosed {
		return fmt.Errorf("channel is already closed")
	}

	if !c.isStarted {
		c.Start()
	}

	c.source <- data

	return nil
}

func (c *rspChannel) RetriveData() <-chan models.Response {
	if c.isClosed {
		ch := make(chan models.Response, 0)
		close(ch)

		return ch
	}

	if !c.isStarted {
		c.Start()
	}

	ch := make(chan models.Response)

	c.retriveLock.Lock()
	c.newDestinations = append(c.newDestinations, ch)
	c.retriveLock.Unlock()

	return ch
}

func (c *rspChannel) Connect(child Channel) error {
	if c.isClosed {
		return fmt.Errorf("channel is already closed")
	}

	if !c.isStarted {
		c.Start()
	}

	c.connectLock.Lock()
	c.newSubChannels = append(c.newSubChannels, child)
	c.connectLock.Unlock()

	child.Start()

	return nil
}

func (c *rspChannel) Start() error {
	if c.isStarted {
		return fmt.Errorf("channel is already started")
	}

	if c.isClosed {
		return fmt.Errorf("channel is already closed")
	}

	c.startOnce.Do(func() {
		if c.source == nil {
			c.source = make(chan models.Response)
		}

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

func (c *rspChannel) Close() error {
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

func (c *rspChannel) mergeNewDest() {
	c.retriveLock.Lock()
	defer func() {
		c.retriveLock.Unlock()
	}()

	if len(c.newDestinations) > 0 {
		c.destinations = append(c.destinations, c.newDestinations...)

		c.newDestinations = c.newDestinations[len(c.newDestinations):]
	}
}

func (c *rspChannel) dispatchDistinations(data models.Response) {
	var invalidDest []int

	c.mergeNewDest()

	for idx, dest := range c.destinations {
		if dest == nil {
			invalidDest = append(invalidDest, idx)
			continue
		}

		dest <- data
	}

	if len(invalidDest) > 0 {
		tmpSlice := make([]interface{}, len(c.destinations))

		for idx, ele := range c.destinations {
			tmpSlice[idx] = ele
		}

		tmpSlice = utils.RangeSlice(tmpSlice, invalidDest)

		c.destinations = make([]chan models.Response, len(tmpSlice))

		for idx, ele := range tmpSlice {
			c.destinations[idx] = ele.(chan models.Response)
		}
	}
}

func (c *rspChannel) mergeNewSubChannel() {
	c.connectLock.Lock()
	defer func() {
		c.connectLock.Unlock()
	}()

	if len(c.newSubChannels) > 0 {
		c.subChannels = append(c.subChannels, c.newSubChannels...)

		c.newSubChannels = c.newSubChannels[len(c.newSubChannels):]
	}
}

func (c *rspChannel) dispatchSubChannels(data models.Response) {
	var invalidSub []int

	for idx, subChan := range c.subChannels {
		if err := subChan.PublishData(data); err != nil {
			invalidSub = append(invalidSub, idx)
			log.Println(err)
			continue
		}

		subChan.PublishData(data)
	}

	if len(invalidSub) > 0 {
		tmpSlice := make([]interface{}, len(c.subChannels))

		for idx, ele := range c.subChannels {
			tmpSlice[idx] = ele
		}

		tmpSlice = utils.RangeSlice(tmpSlice, invalidSub)

		c.subChannels = make([]Channel, len(tmpSlice))

		for idx, ele := range tmpSlice {
			c.subChannels[idx] = ele.(Channel)
		}
	}
}
