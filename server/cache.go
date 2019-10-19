package server

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/frozenpine/wstester/models"
)

// CacheLimitType limited cache types
type CacheLimitType int

const (
	// Realtime get all real time mbl notify
	Realtime CacheLimitType = iota
	// Realtime25 get real time mbl notify under depth 25
	Realtime25
	// Depth25 get real time mbl in depth25 snapshot
	Depth25
	// Quote get quote tick in 500ms
	Quote
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
	TakeSnapshot(depth int, publish Channel) models.TableResponse
	// Append append data to cache
	// this is an async operation if cache pipeline not full.
	Append(in *CacheInput)
	// GetLimitedChannel get limited response channel
	GetLimitedRsp(limit CacheLimitType) Channel
}

// CacheInput wrapper structure for table response
type CacheInput struct {
	pubToDefault   bool
	pubChannel     Channel
	breakpointFunc func() models.TableResponse
	msg            []byte
}

// IsBreakPoint to check input is a breakpoint message
func (in *CacheInput) IsBreakPoint() bool {
	return in.breakpointFunc != nil
}

// NewCacheInput make a new cache input
func NewCacheInput(msg []byte) *CacheInput {
	cache := CacheInput{
		pubToDefault: true,
		msg:          msg,
	}

	return &cache
}

type tableCache struct {
	rspChannel

	pipeline  chan *CacheInput
	ready     chan struct{}
	startOnce sync.Once
	IsReady   bool
	IsClosed  bool
	stopOnce  sync.Once
	maxLength int

	snapshotFn    func(int) models.TableResponse
	handleInputFn func(*CacheInput)
}

func (c *tableCache) Start() error {
	if c.isStarted {
		return errors.New("cache is already started")
	}

	if c.isClosed {
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

		close(c.ready)
		c.IsReady = true
		c.IsClosed = false

		if !c.rspChannel.isStarted {
			c.rspChannel.Start()
		}
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

		if !c.rspChannel.isClosed {
			c.rspChannel.Close()
		}
	})

	return nil
}

func (c *tableCache) Ready() <-chan struct{} {
	return c.ready
}

func (c *tableCache) TakeSnapshot(depth int, publish Channel) models.TableResponse {
	ch := make(chan models.TableResponse, 1)

	c.pipeline <- &CacheInput{
		pubChannel: publish,
		breakpointFunc: func() models.TableResponse {
			if c.snapshotFn == nil {
				log.Panicln("snapshotFn is nil.")
			}

			snap := c.snapshotFn(depth)

			ch <- snap
			close(ch)

			return snap
		},
	}

	return <-ch
}

func (c *tableCache) GetLimitedRsp(limit CacheLimitType) Channel {
	return c
}

func (c *tableCache) Append(in *CacheInput) {
	c.pipeline <- in
}
