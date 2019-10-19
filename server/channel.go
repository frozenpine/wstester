package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/frozenpine/wstester/models"
	"github.com/frozenpine/wstester/utils"
)

// Channel message channel
type Channel interface {
	// Start initialize channel and start a dispatch goroutine
	Start() error
	// Close close channel input
	Close() error
	// Connect connect child channel, child channel will get dispatched data from current channel
	Connect(subChan Channel) error
	// TODO: Disconnect sub channel
	// Disconnect(subChan Channel) error
	// PublishData publish data to current channel
	PublishData(rsp models.Response) error
	// RetriveData to get an chan to retrive data in current channel
	RetriveData(client Session) <-chan models.Response
}

type rspChannel struct {
	source chan models.Response

	destinations     map[Session]chan models.Response
	newDestinations  map[Session]chan models.Response
	childChannels    []Channel
	newChildChannels []Channel
	retriveLock      sync.Mutex
	connectLock      sync.Mutex

	ctx       context.Context
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

func (c *rspChannel) RetriveData(client Session) <-chan models.Response {
	if c.isClosed {
		ch := make(chan models.Response, 0)
		close(ch)

		return ch
	}

	if !c.isStarted {
		c.Start()
	}

	ch := make(chan models.Response, 1000)

	c.retriveLock.Lock()
	c.newDestinations[client] = ch
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
	c.newChildChannels = append(c.newChildChannels, child)
	c.connectLock.Unlock()

	child.Start()

	return nil
}

func (c *rspChannel) Start() error {
	if c.isStarted {
		return errors.New("channel is already started")
	}

	if c.isClosed {
		return errors.New("channel is already closed")
	}

	c.startOnce.Do(func() {
		if c.source == nil {
			c.source = make(chan models.Response, 1000)
		}

		if c.destinations == nil {
			c.destinations = make(map[Session]chan models.Response)
		}

		if c.newDestinations == nil {
			c.newDestinations = make(map[Session]chan models.Response)
		}

		c.isClosed = false

		go func() {
			defer c.Close()

			for {
				select {
				case <-c.ctx.Done():
					return
				case data := <-c.source:
					if data == nil {
						continue
					}

					c.dispatchDistinations(data)
					c.dispatchSubChannels(data)
				}
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

		for _, ch := range c.destinations {
			close(ch)
		}

		log.Println("channel closed.")
	})

	return nil
}

func (c *rspChannel) mergeNewDestinations() {
	c.retriveLock.Lock()
	defer c.retriveLock.Unlock()

	var merged []Session

	for client, dest := range c.newDestinations {
		c.destinations[client] = dest
		merged = append(merged, client)
	}

	for _, client := range merged {
		delete(c.newDestinations, client)
	}
}

func (c *rspChannel) dispatchDistinations(data models.Response) {
	var invalidDest []Session

	c.mergeNewDestinations()

	writeTimeout := time.NewTimer(time.Second * 5)

	for client, dest := range c.destinations {
		if client.IsClosed() {
			invalidDest = append(invalidDest, client)
			writeTimeout.Reset(time.Second * 5)
			continue
		}

		select {
		case dest <- data:
			writeTimeout.Reset(time.Second * 5)
		case <-writeTimeout.C:
			invalidDest = append(invalidDest, client)
			writeTimeout = time.NewTimer(time.Second * 5)
			log.Printf("Dispatch data to client[%s] timeout.", client.GetID())
		}
	}

	writeTimeout.Stop()

	for _, closedClient := range invalidDest {
		delete(c.destinations, closedClient)
	}
}

func (c *rspChannel) mergeNewSubChannel() {
	c.connectLock.Lock()
	defer c.connectLock.Unlock()

	if len(c.newChildChannels) > 0 {
		c.childChannels = append(c.childChannels, c.newChildChannels...)

		c.newChildChannels = c.newChildChannels[len(c.newChildChannels):]
	}
}

func (c *rspChannel) dispatchSubChannels(data models.Response) {
	var invalidSub []int

	c.mergeNewSubChannel()

	for idx, subChan := range c.childChannels {
		if err := subChan.PublishData(data); err != nil {
			invalidSub = append(invalidSub, idx)
			log.Println(err)
			continue
		}
	}

	if len(invalidSub) > 0 {
		tmpSlice := make([]interface{}, len(c.childChannels))

		for idx, ele := range c.childChannels {
			tmpSlice[idx] = ele
		}

		tmpSlice = utils.RangeSlice(tmpSlice, invalidSub)

		c.childChannels = make([]Channel, len(tmpSlice))

		for idx, ele := range tmpSlice {
			c.childChannels[idx] = ele.(Channel)
		}
	}
}
