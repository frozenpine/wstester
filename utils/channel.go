package utils

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/frozenpine/wstester/models"
)

const (
	dispatchTimeout = 5
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
	PublishData(rsp models.TableResponse) error
	// RetriveData to get an chan to retrive data in current channel
	RetriveData() <-chan models.TableResponse
}

type rspChannel struct {
	source chan models.TableResponse

	destinations     []chan<- models.TableResponse
	newDestinations  []chan<- models.TableResponse
	childChannels    []Channel
	newChildChannels []Channel
	retriveLock      sync.Mutex
	connectLock      sync.Mutex

	ctx      context.Context
	IsReady  bool
	IsClosed bool
}

func (c *rspChannel) PublishData(data models.TableResponse) error {
	if c.IsClosed {
		return fmt.Errorf("channel is already closed")
	}

	// if !c.IsReady {
	// 	c.Start()
	// }

	c.source <- data

	return nil
}

func (c *rspChannel) RetriveData() <-chan models.TableResponse {
	if c.IsClosed {
		return nil
	}

	// if !c.IsReady {
	// 	c.Start()
	// }

	ch := make(chan models.TableResponse, 1)

	c.retriveLock.Lock()
	c.newDestinations = append(c.newDestinations, ch)
	c.retriveLock.Unlock()

	return ch
}

func (c *rspChannel) Connect(child Channel) error {
	if c.IsClosed {
		return fmt.Errorf("channel is already closed")
	}

	if !c.IsReady {
		c.Start()
	}

	c.connectLock.Lock()
	c.newChildChannels = append(c.newChildChannels, child)
	c.connectLock.Unlock()

	child.Start()

	return nil
}

func (c *rspChannel) Start() error {
	if c.IsReady {
		return errors.New("channel is already started")
	}

	if c.IsClosed {
		return errors.New("channel is already closed")
	}

	if c.source == nil {
		c.source = make(chan models.TableResponse, 1000)
	}

	c.IsClosed = false

	go func() {
		defer c.Close()

		for {
			select {
			case <-c.ctx.Done():
				return
			case data, ok := <-c.source:
				if !ok {
					return
				}

				if data == nil {
					continue
				}

				c.dispatchDistinations(data)
				c.dispatchSubChannels(data)
			}
		}
	}()

	c.IsReady = true

	return nil
}

func (c *rspChannel) Close() error {
	if !c.IsReady {
		return fmt.Errorf("channel is not started")
	}

	if c.IsClosed {
		return fmt.Errorf("channel is already closed")
	}

	c.IsReady = false

	close(c.source)

	c.IsClosed = true

	for _, ch := range c.destinations {
		close(ch)
	}

	log.Println("channel closed.")

	return nil
}

func (c *rspChannel) mergeNewDestinations() {
	c.retriveLock.Lock()
	defer c.retriveLock.Unlock()

	if len(c.newDestinations) > 0 {
		c.destinations = append(c.destinations, c.newDestinations...)

		c.newDestinations = []chan<- models.TableResponse{}
	}
}

func (c *rspChannel) dispatchDistinations(data models.TableResponse) {
	var invalidDest []int

	c.mergeNewDestinations()

	writeTimeout := time.NewTimer(time.Second * dispatchTimeout)

	for idx, dest := range c.destinations {
		if dest == nil {
			invalidDest = append(invalidDest, idx)
			log.Println("Destination channel is nil")
			continue
		}

		select {
		case dest <- data:
			writeTimeout.Reset(time.Second * dispatchTimeout)
		case <-writeTimeout.C:
			close(dest)
			invalidDest = append(invalidDest, idx)
			writeTimeout = time.NewTimer(time.Second * dispatchTimeout)
			log.Printf("Dispatch data to client timeout.")
		}
	}

	writeTimeout.Stop()

	// FIXME: 可能存在未正确处理的destination
	if len(invalidDest) > 0 {
		tmpSlice := make([]interface{}, len(c.destinations))

		for idx, ele := range c.destinations {
			tmpSlice[idx] = ele
		}

		tmpSlice = RangeSlice(tmpSlice, invalidDest)

		c.destinations = make([]chan<- models.TableResponse, len(tmpSlice))

		for idx, ele := range tmpSlice {
			c.destinations[idx] = ele.(chan<- models.TableResponse)
		}
	}
}

func (c *rspChannel) mergeNewSubChannel() {
	c.connectLock.Lock()
	defer c.connectLock.Unlock()

	if len(c.newChildChannels) > 0 {
		c.childChannels = append(c.childChannels, c.newChildChannels...)

		c.newChildChannels = []Channel{}
	}
}

func (c *rspChannel) dispatchSubChannels(data models.TableResponse) {
	var invalidSub []int

	c.mergeNewSubChannel()

	for idx, subChan := range c.childChannels {
		if err := subChan.PublishData(data); err != nil {
			subChan.Close()
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

		tmpSlice = RangeSlice(tmpSlice, invalidSub)

		c.childChannels = make([]Channel, len(tmpSlice))

		for idx, ele := range tmpSlice {
			c.childChannels[idx] = ele.(Channel)
		}
	}
}
