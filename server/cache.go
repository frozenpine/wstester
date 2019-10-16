package server

import (
	"context"
	"log"

	"github.com/frozenpine/wstester/models"
)

// Cache cache for table response
type Cache interface {
	Start(context.Context) error
	Stop() error
	Ready() <-chan struct{}
	TakeSnapshot(bool) models.TableResponse
	Append(*CacheInput)
}

// CacheInput wrapper structure for table response
type CacheInput struct {
	pubToChannel   bool
	breakpointFunc func() models.TableResponse
	msg            []byte
}

// NewCacheInput make a new cache input
func NewCacheInput(msg []byte) *CacheInput {
	cache := CacheInput{
		pubToChannel: true,
		msg:          msg,
	}

	return &cache
}

// IsBreakPoint to check input is a breakpoint message
func (in *CacheInput) IsBreakPoint() bool {
	return in.breakpointFunc != nil
}

type tableCache struct {
	rspChannel

	pipeline  chan *CacheInput
	ready     chan struct{}
	IsReady   bool
	close     chan struct{}
	IsClosed  bool
	maxLength int

	snapshotFn    func() models.TableResponse
	handleInputFn func(*CacheInput) models.TableResponse
}

func (c *tableCache) Start(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	c.ctx = ctx

	if c.pipeline == nil {
		c.pipeline = make(chan *CacheInput)
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
			case <-c.close:
				return
			case obj := <-c.pipeline:
				if obj == nil {
					continue
				}

				if c.handleInputFn == nil {
					log.Println("handleInputFn is nil.")

					continue
				}

				rsp := c.handleInputFn(obj)

				if rsp != nil && obj.pubToChannel {
					c.PublishData(rsp)
				}
			}
		}
	}()

	c.IsReady = true
	c.IsClosed = false
	close(c.ready)

	return nil
}

func (c *tableCache) Stop() error {
	close(c.close)
	return nil
}

func (c *tableCache) Ready() <-chan struct{} {
	return c.ready
}

// TakeSnapshot to get snapshot of cache, if publish is true, snapshot result will be notified in channel
func (c *tableCache) TakeSnapshot(publish bool) models.TableResponse {
	ch := make(chan models.TableResponse, 1)
	defer func() {
		close(ch)
	}()

	c.pipeline <- &CacheInput{
		pubToChannel: publish,
		breakpointFunc: func() models.TableResponse {
			if c.snapshotFn == nil {
				log.Println("snapshotFn is nil.")

				return nil
			}

			snap := c.snapshotFn()

			ch <- snap

			return snap
		},
	}

	return <-ch
}

func (c *tableCache) Append(in *CacheInput) {
	c.pipeline <- in
}
