package client

import (
	"context"
	"errors"
	"log"
	"sync"

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
	TakeSnapshot(depth int, publish ...chan<- models.TableResponse) models.TableResponse
	// GetRspChannel get response channel
	GetRspChannel() <-chan models.TableResponse
	// Attatch attatch cache input to an channel
	Attatch(<-chan models.TableResponse)
}

// CacheInput wrapper structure for table response
type CacheInput struct {
	pubChannels    []chan<- models.TableResponse
	breakpointFunc func() models.TableResponse
	msg            models.TableResponse
}

// IsBreakPoint to check input is a breakpoint message
func (in *CacheInput) IsBreakPoint() bool {
	return in.breakpointFunc != nil
}

// NewCacheInput make a new cache input
func NewCacheInput(rsp models.TableResponse) *CacheInput {
	input := CacheInput{
		msg: rsp,
	}

	return &input
}

// NewBreakpoint make a new cache breakpoint
func NewBreakpoint(breakpointFn func() models.TableResponse, publish ...chan<- models.TableResponse) *CacheInput {
	input := CacheInput{
		pubChannels:    publish,
		breakpointFunc: breakpointFn,
	}

	return &input
}

type tableCache struct {
	output chan models.TableResponse

	pipeline  chan *CacheInput
	ready     chan struct{}
	ctx       context.Context
	startOnce sync.Once
	IsReady   bool
	IsClosed  bool
	stopOnce  sync.Once
	maxLength int

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

	c.startOnce.Do(func() {
		if c.ctx == nil {
			c.ctx = context.Background()
		}

		if c.pipeline == nil {
			c.pipeline = make(chan *CacheInput, 1000)
		}

		if c.ready == nil {
			c.ready = make(chan struct{})
		}

		c.output = make(chan models.TableResponse)

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
	})

	return nil
}

func (c *tableCache) Stop() error {
	if c.IsClosed == true {
		return errors.New("cache is already stopped")
	}
	if c.IsReady == false {
		return errors.New("cache is not ready")
	}

	c.stopOnce.Do(func() {
		c.IsClosed = true
		c.IsReady = false
		close(c.pipeline)
		close(c.output)
	})

	return nil
}

func (c *tableCache) Ready() <-chan struct{} {
	return c.ready
}

func (c *tableCache) TakeSnapshot(depth int, publish ...chan<- models.TableResponse) models.TableResponse {
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

func (c *tableCache) GetRspChannel() <-chan models.TableResponse {
	return c.output
}

func (c *tableCache) Attatch(ch <-chan models.TableResponse) {
	if ch == nil {
		return
	}

	go func() {
		for rsp := range ch {
			c.pipeline <- NewCacheInput(rsp)
		}
	}()
}
