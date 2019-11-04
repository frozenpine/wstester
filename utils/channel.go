package utils

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/frozenpine/wstester/models"
	"github.com/frozenpine/wstester/utils/log"
	uuid "github.com/satori/go.uuid"
)

const (
	dispatchTimeout = 3
)

// Input cache & channel input
type Input interface {
	IsBreakpoint() bool
	GetData() models.TableResponse
}

// ChannelInput channel input structure
type ChannelInput struct {
	breakpointFunc             func()
	dstSession, subChanSession string
	rsp                        models.TableResponse
}

// IsBreakpoint to check if input is a breakpoint message
func (in *ChannelInput) IsBreakpoint() bool {
	return in.breakpointFunc != nil
}

// GetData get input data
func (in *ChannelInput) GetData() models.TableResponse {
	return in.rsp
}

// NewChannelInput make a new channel input
func NewChannelInput(rsp models.TableResponse) *ChannelInput {
	input := ChannelInput{
		rsp: rsp,
	}

	return &input
}

// NewChannelBreakpoint make a new channel breakpoint
func NewChannelBreakpoint(fn func()) *ChannelInput {
	input := ChannelInput{
		breakpointFunc: fn,
	}

	return &input
}

// Channel message channel
type Channel interface {
	// Start initialize channel and start a dispatch goroutine
	Start() error

	// Close close channel input
	Close() error

	// Connect connect child channel, child channel will get dispatched data from current channel
	Connect(subChan Channel) (string, error)

	// Disconnect disconnect child channel
	Disconnect(session string) error

	// PublishData publish data to current channel
	PublishData(rsp models.TableResponse) error

	// PublishDataToDestination publish data to specified destination
	PublishDataToDestination(rsp models.TableResponse, session string) error

	// PublishDataToSubChan publish data to specified sub channel
	PublishDataToSubChan(rsp models.TableResponse, session string) error

	// RetriveData to get an chan to retrive data in current channel
	RetriveData() (string, <-chan models.TableResponse)

	// ShutdownRetrive shutdown data chan specified by session
	ShutdownRetrive(session string) error
}

type rspChannel struct {
	source chan *ChannelInput

	destinations  map[string]chan<- models.TableResponse
	childChannels map[string]Channel

	ctx      context.Context
	IsReady  bool
	IsClosed bool
}

func (c *rspChannel) PublishData(data models.TableResponse) error {
	if c.IsClosed {
		return fmt.Errorf("channel is already closed")
	}

	c.source <- &ChannelInput{
		rsp: data,
	}

	return nil
}

func (c *rspChannel) PublishDataToDestination(data models.TableResponse, session string) error {
	if c.IsClosed {
		return fmt.Errorf("channel is already closed")
	}

	c.source <- &ChannelInput{
		dstSession: session,
		rsp:        data,
	}

	return nil
}

func (c *rspChannel) PublishDataToSubChan(data models.TableResponse, session string) error {
	if c.IsClosed {
		return fmt.Errorf("channel is already closed")
	}

	c.source <- &ChannelInput{
		subChanSession: session,
		rsp:            data,
	}

	return nil
}

func (c *rspChannel) RetriveData() (string, <-chan models.TableResponse) {
	if c.IsClosed {
		return "", nil
	}

	ch := make(chan models.TableResponse, 1000)
	session := uuid.NewV4().String()

	c.source <- NewChannelBreakpoint(func() {
		c.destinations[session] = ch
	})

	return session, ch
}

func (c *rspChannel) ShutdownRetrive(session string) error {
	if c.IsClosed {
		return nil
	}

	ch := make(chan error, 1)

	c.source <- NewChannelBreakpoint(func() {
		if dst, exist := c.destinations[session]; exist {
			close(dst)
			ch <- nil
		} else {
			ch <- fmt.Errorf("destination session[%s] not exists", session)
		}

		close(ch)
	})

	return <-ch
}

func (c *rspChannel) Connect(child Channel) (string, error) {
	if c.IsClosed {
		return "", fmt.Errorf("channel is already closed")
	}

	if !c.IsReady {
		c.Start()
	}

	session := uuid.NewV4().String()

	c.source <- NewChannelBreakpoint(func() {
		c.childChannels[session] = child
	})

	return session, nil
}

func (c *rspChannel) Disconnect(session string) error {
	if c.IsClosed {
		return nil
	}

	ch := make(chan error, 1)

	c.source <- NewChannelBreakpoint(func() {
		if _, exist := c.childChannels[session]; exist {
			delete(c.childChannels, session)
			ch <- nil
		} else {
			ch <- fmt.Errorf("invalid sub channel session[%s]", session)
		}
		close(ch)
	})

	return <-ch
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

				if input.IsBreakpoint() {
					input.breakpointFunc()
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

func (c *rspChannel) dispatchDistinations(data *ChannelInput) {
	var invalidDest []string

	handleInput := func(session string, dest chan<- models.TableResponse, writeTimeout *time.Timer) {
		if dest == nil {
			invalidDest = append(invalidDest, session)
			log.Error("Destination channel is nil")
			return
		}

		select {
		case dest <- data.rsp:
			writeTimeout.Reset(time.Second * dispatchTimeout)
		case <-writeTimeout.C:
			close(dest)
			invalidDest = append(invalidDest, session)
			writeTimeout = time.NewTimer(time.Second * dispatchTimeout)
			log.Warn("Dispatch data to client timeout.")
		}
	}

	writeTimeout := time.NewTimer(time.Second * dispatchTimeout)
	if data.dstSession == "" {
		for session, dest := range c.destinations {
			handleInput(session, dest, writeTimeout)
		}
	} else {
		if dst, exist := c.destinations[data.dstSession]; exist {
			handleInput(data.dstSession, dst, writeTimeout)
		} else {
			log.Errorf("Invalid destination session[%s] specified", data.dstSession)
		}
	}
	writeTimeout.Stop()

	if len(invalidDest) > 0 {
		for _, invalid := range invalidDest {
			delete(c.destinations, invalid)
		}
	}
}

func (c *rspChannel) dispatchSubChannels(data *ChannelInput) {
	var invalidSub []string

	handleInput := func(session string, subChan Channel) {
		if err := subChan.PublishData(data.rsp); err != nil {
			subChan.Close()
			invalidSub = append(invalidSub, session)
			log.Error(err)
		}
	}

	if data.subChanSession == "" {
		for session, subChan := range c.childChannels {
			handleInput(session, subChan)
		}
	} else {
		if sub, exist := c.childChannels[data.subChanSession]; exist {
			handleInput(data.subChanSession, sub)
		} else {
			log.Errorf("Invalid sub channel session[%s]", data.dstSession)
		}
	}

	if len(invalidSub) > 0 {
		for _, invalid := range invalidSub {
			delete(c.childChannels, invalid)
		}
	}
}
