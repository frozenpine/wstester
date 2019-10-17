package server

import (
	"context"
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
	Start(ctx context.Context) error
	// Close close channel
	Close() error
	// Connect connect child channel, child channel will get dispatched data from current channel
	Connect(subChan Channel) error

	// PublishData publish data to current channel
	PublishData(rsp models.Response) error
	// RetriveData to get an chan to retrive data in current channel
	RetriveData(client Session) <-chan models.Response
}

type rspChannel struct {
	source chan models.Response

	destinations     map[Session]chan models.Response
	childChannels    []Channel
	newChildChannels []Channel
	retriveLock      sync.RWMutex
	connectLock      sync.RWMutex

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
	c.destinations[client] = ch
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

	child.Start(c.ctx)

	return nil
}

func (c *rspChannel) cleanup() {
	for _, ch := range c.destinations {
		close(ch)
	}

	for _, subChan := range c.childChannels {
		subChan.Close()
	}

	c.isStarted = false
	c.isClosed = true
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
			c.source = make(chan models.Response, 1000)
		}

		c.isClosed = false

		go func() {
			defer c.cleanup()

			for {
				select {
				case <-c.ctx.Done():
					log.Println("channel closed.")
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
	})

	return nil
}

func (c *rspChannel) dispatchDistinations(data models.Response) {
	var invalidDest []Session

	writeTimeout := time.NewTimer(time.Second * 5)

	c.retriveLock.RLock()
	for client, dest := range c.destinations {
		if client.IsClosed() {
			invalidDest = append(invalidDest, client)
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
	c.retriveLock.RUnlock()

	writeTimeout.Stop()

	c.retriveLock.Lock()
	for _, closedClient := range invalidDest {
		delete(c.destinations, closedClient)
	}
	c.retriveLock.Unlock()
}

func (c *rspChannel) mergeNewSubChannel() {
	c.connectLock.Lock()
	defer func() {
		c.connectLock.Unlock()
	}()

	if len(c.newChildChannels) > 0 {
		c.childChannels = append(c.childChannels, c.newChildChannels...)

		c.newChildChannels = c.newChildChannels[len(c.newChildChannels):]
	}
}

func (c *rspChannel) dispatchSubChannels(data models.Response) {
	var invalidSub []int

	for idx, subChan := range c.childChannels {
		if err := subChan.PublishData(data); err != nil {
			invalidSub = append(invalidSub, idx)
			log.Println(err)
			continue
		}

		subChan.PublishData(data)
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
