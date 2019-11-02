package utils

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/frozenpine/wstester/models"
	"github.com/frozenpine/wstester/utils/log"
)

const (
	dispatchTimeout = 3
)

// ChannelInput channel input structure
type ChannelInput struct {
	dstIdx, childChanIdx int
	rsp                  models.TableResponse
}

// Channel message channel
type Channel interface {
	// Start initialize channel and start a dispatch goroutine
	Start() error

	// Close close channel input
	Close() error

	// Connect connect child channel, child channel will get dispatched data from current channel
	Connect(subChan Channel) (int, error)

	// TODO: Disconnect sub channel
	// Disconnect(subChan Channel) error

	// PublishData publish data to current channel
	PublishData(rsp models.TableResponse) error

	// PublishDataToDestination publish data to specified destination
	PublishDataToDestination(rsp models.TableResponse, idx int) error

	// PublishDataToSubChan publish data to specified sub channel
	PublishDataToSubChan(rsp models.TableResponse, idx int) error

	// RetriveData to get an chan to retrive data in current channel
	RetriveData() (int, <-chan models.TableResponse)
}

type rspChannel struct {
	source chan *ChannelInput

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

	c.source <- &ChannelInput{
		dstIdx:       -1,
		childChanIdx: -1,
		rsp:          data,
	}

	return nil
}

func (c *rspChannel) PublishDataToDestination(data models.TableResponse, idx int) error {
	if c.IsClosed {
		return fmt.Errorf("channel is already closed")
	}

	c.source <- &ChannelInput{
		dstIdx:       idx,
		childChanIdx: -1,
		rsp:          data,
	}

	return nil
}

func (c *rspChannel) PublishDataToSubChan(data models.TableResponse, idx int) error {
	if c.IsClosed {
		return fmt.Errorf("channel is already closed")
	}

	c.source <- &ChannelInput{
		dstIdx:       -1,
		childChanIdx: idx,
		rsp:          data,
	}

	return nil
}

func (c *rspChannel) RetriveData() (int, <-chan models.TableResponse) {
	if c.IsClosed {
		return -1, nil
	}

	ch := make(chan models.TableResponse, 1000)

	c.retriveLock.Lock()
	c.newDestinations = append(c.newDestinations, ch)
	idx := len(c.newDestinations) - 1 + len(c.destinations)
	c.retriveLock.Unlock()

	return idx, ch
}

func (c *rspChannel) Connect(child Channel) (int, error) {
	if c.IsClosed {
		return -1, fmt.Errorf("channel is already closed")
	}

	if !c.IsReady {
		c.Start()
	}

	c.connectLock.Lock()
	c.newChildChannels = append(c.newChildChannels, child)
	idx := len(c.newChildChannels) - 1 + len(c.childChannels)
	c.connectLock.Unlock()

	child.Start()

	return idx, nil
}

func (c *rspChannel) Start() error {
	if c.IsReady {
		return errors.New("channel is already started")
	}

	if c.IsClosed {
		return errors.New("channel is already closed")
	}

	if c.source == nil {
		c.source = make(chan *ChannelInput, 1000)
	}

	c.IsClosed = false

	go func() {
		defer c.Close()

		for {
			select {
			case <-c.ctx.Done():
				return
			case input, ok := <-c.source:
				if !ok {
					return
				}

				if input == nil {
					continue
				}

				c.dispatchDistinations(input)
				c.dispatchSubChannels(input)
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

	return nil
}

func (c *rspChannel) mergeNewDestinations() {
	if len(c.newDestinations) < 1 {
		return
	}

	c.retriveLock.Lock()

	c.destinations = append(c.destinations, c.newDestinations...)
	c.newDestinations = []chan<- models.TableResponse{}

	c.retriveLock.Unlock()
}

func (c *rspChannel) dispatchDistinations(data *ChannelInput) {
	var invalidDest []int

	c.mergeNewDestinations()

	handleInput := func(idx int, dest chan<- models.TableResponse, writeTimeout *time.Timer) {
		if dest == nil {
			invalidDest = append(invalidDest, idx)
			log.Error("Destination channel is nil")
			return
		}

		// TODO： 更好的检测目标chan关闭的机制
		select {
		case dest <- data.rsp:
			writeTimeout.Reset(time.Second * dispatchTimeout)
		case <-writeTimeout.C:
			close(dest)
			invalidDest = append(invalidDest, idx)
			writeTimeout = time.NewTimer(time.Second * dispatchTimeout)
			log.Warn("Dispatch data to client timeout.")
		}
	}

	writeTimeout := time.NewTimer(time.Second * dispatchTimeout)
	if data.dstIdx < 0 {
		for idx, dest := range c.destinations {
			handleInput(idx, dest, writeTimeout)
		}
	} else {
		if data.dstIdx < len(c.destinations) {
			handleInput(data.dstIdx, c.destinations[data.dstIdx], writeTimeout)
		} else {
			log.Errorf("Invalid destination index[%d] specified, max index is %d", data.dstIdx, len(c.destinations)-1)
		}
	}
	writeTimeout.Stop()

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
	if len(c.newChildChannels) < 1 {
		return
	}

	c.connectLock.Lock()

	c.childChannels = append(c.childChannels, c.newChildChannels...)
	c.newChildChannels = []Channel{}

	c.connectLock.Unlock()
}

func (c *rspChannel) dispatchSubChannels(data *ChannelInput) {
	var invalidSub []int

	c.mergeNewSubChannel()

	handleInput := func(idx int, subChan Channel) {
		if err := subChan.PublishData(data.rsp); err != nil {
			subChan.Close()
			invalidSub = append(invalidSub, idx)
			log.Error(err)
		}
	}

	if data.childChanIdx < 0 {
		for idx, subChan := range c.childChannels {
			handleInput(idx, subChan)
		}
	} else {
		if data.childChanIdx < len(c.childChannels) {
			handleInput(data.childChanIdx, c.childChannels[data.childChanIdx])
		} else {
			log.Errorf("Invalid sub channel index[%d] specified, max index is %d\n",
				data.dstIdx, len(c.childChannels)-1)
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
