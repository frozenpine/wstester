package utils

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/frozenpine/wstester/models"
)

// ChannelType limited cache types
type ChannelType int

const (
	// Realtime get all real time mbl notify
	Realtime ChannelType = iota
	// Snapshot get real time mbl notify in snapshot
	Snapshot
	// Tick get tick mbl notify in depth25 snapshot
	Tick
)

// Cache cache for table response
type Cache interface {
	// Start start cache backgroud loop.
	Start() error
	// Stop stop cache backgroud loop.
	Stop() error
	// Ready wait for cache ready.
	Ready() <-chan struct{}
	// TakeSnapshot take snapshot for cache,
	// depth <= 0 means all available depth level,
	// publish means wether publish snapshot in channel,
	// this is an async to sync operation, snapshot operation queued in cache pipeline and
	// return util queued operation finished.
	TakeSnapshot(depth int, publish ...Channel) models.TableResponse
	// Append append data to cache
	// this is an async operation if cache pipeline not full.
	Append(in *CacheInput)
	// GetRspChannel get response channel
	GetRspChannel(chType ChannelType, depth int) Channel
	// GetDefaultChannel get default channel with realtime all depth notify
	GetDefaultChannel() Channel
}

// CacheInput wrapper structure for table response
type CacheInput struct {
	pubChannels    []Channel
	breakpointFunc func() models.TableResponse
	msg            models.TableResponse
}

// IsBreakPoint to check input is a breakpoint message
func (in *CacheInput) IsBreakPoint() bool {
	return in.breakpointFunc != nil
}

// NewCacheInput make a new cache input
func NewCacheInput(msg models.TableResponse) *CacheInput {
	input := CacheInput{
		msg: msg,
	}

	return &input
}

// NewBreakpoint make a new cache breakpoint
func NewBreakpoint(breakpointFn func() models.TableResponse, publish ...Channel) *CacheInput {
	input := CacheInput{
		pubChannels:    publish,
		breakpointFunc: breakpointFn,
	}

	return &input
}

type tableCache struct {
	channelGroup [3]map[int]Channel

	Symbol     string
	pipeline   chan *CacheInput
	cacheStart time.Time
	ready      chan struct{}
	ctx        context.Context
	IsReady    bool
	IsClosed   bool
	maxLength  int

	snapshotFn    func(int) models.TableResponse
	handleInputFn func(*CacheInput)
}

func (c *tableCache) Start() error {
	if c.IsReady {
		return errors.New("cache is already started")
	}

	if c.IsClosed {
		return errors.New("cache is already closed")
	}

	for _, chGroup := range c.channelGroup {
		if chGroup == nil {
			continue
		}

		for _, rspChan := range chGroup {
			rspChan.Start()
		}
	}

	go func() {
		defer func() {
			c.IsReady = false
			c.IsClosed = true
		}()

		for {
			select {
			case <-c.ctx.Done():
				return
			case obj := <-c.pipeline:
				if obj == nil {
					continue
				}

				if c.handleInputFn == nil {
					log.Panicln("handleInputFn is nil.")
				}

				c.handleInputFn(obj)
			}
		}
	}()

	c.IsReady = true
	c.IsClosed = false
	close(c.ready)
	c.cacheStart = time.Now()

	return nil
}

func (c *tableCache) Stop() error {
	if c.IsClosed == true {
		return errors.New("cache is already stopped")
	}
	if c.IsReady == false {
		return errors.New("cache is not ready")
	}

	c.IsClosed = true
	c.IsReady = false
	close(c.pipeline)

	for _, chGroup := range c.channelGroup {
		for _, rspChan := range chGroup {
			rspChan.Close()
		}
	}

	return nil
}

func (c *tableCache) Ready() <-chan struct{} {
	return c.ready
}

func (c *tableCache) TakeSnapshot(depth int, publish ...Channel) models.TableResponse {
	ch := make(chan models.TableResponse, 1)

	snapFn := func() models.TableResponse {
		if c.snapshotFn == nil {
			log.Panicln("snapshotFn is nil.")
		}

		snap := c.snapshotFn(depth)

		ch <- snap
		close(ch)

		return snap
	}

	c.pipeline <- NewBreakpoint(snapFn, publish...)

	return <-ch
}

func (c *tableCache) GetRspChannel(chType ChannelType, depth int) Channel {
	if depth < 1 {
		depth = 0
	}

	if chGroup := c.channelGroup[chType]; chGroup != nil {
		return chGroup[depth]
	}

	return nil
}

func (c *tableCache) Append(in *CacheInput) {
	c.pipeline <- in
}

func (c *tableCache) GetDefaultChannel() Channel {
	if !c.IsReady || c.IsClosed {
		return nil
	}

	return c.channelGroup[Realtime][0]
}
