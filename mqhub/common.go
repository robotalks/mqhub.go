package mqhub

import (
	"context"
)

// MessageSinkFunc is func form of MessageSink
type MessageSinkFunc func(Message) Future

// ConsumeMessage implements MessageSink
func (f MessageSinkFunc) ConsumeMessage(msg Message) Future {
	future := f(msg)
	if future == nil {
		future = &ImmediateFuture{}
	}
	return future
}

// DataPoint implements Endpoint for a data point
type DataPoint struct {
	Name   string
	Retain bool
	Sink   MessageSink
}

// NewDataPoint creates a new datapoint
func NewDataPoint(name string) *DataPoint {
	return &DataPoint{Name: name}
}

// NewRetainDataPoint creates a new retain datapoint
func NewRetainDataPoint(name string) *DataPoint {
	return &DataPoint{Name: name, Retain: true}
}

// ID implements Endpoint
func (p *DataPoint) ID() string {
	return p.Name
}

// SinkMessage implements MessageSource
func (p *DataPoint) SinkMessage(sink MessageSink) {
	p.Sink = sink
}

// Update updates the state
func (p *DataPoint) Update(state interface{}) Future {
	sink := p.Sink
	if sink == nil {
		return &ImmediateFuture{Error: ErrNoMessageSink}
	}
	msg, ok := state.(Message)
	if !ok {
		msg = MakeMsg(state, p.Retain)
	}
	return p.Sink.ConsumeMessage(msg)
}

// Reactor implements Endpoint for a reactor to an update
type Reactor struct {
	Name    string
	Handler MessageSink
}

// ReactorFunc creates a Reactor from MessageSinkFunc
func ReactorFunc(name string, handler MessageSinkFunc) *Reactor {
	return &Reactor{Name: name, Handler: handler}
}

// ID implements Endpoint
func (a *Reactor) ID() string {
	return a.Name
}

// ConsumeMessage implements MessageSink
func (a *Reactor) ConsumeMessage(msg Message) Future {
	return a.Handler.ConsumeMessage(msg)
}

// Do sets the message handler
func (a *Reactor) Do(sink MessageSink) *Reactor {
	a.Handler = sink
	return a
}

// DoFunc is same as Do but accepts a func
func (a *Reactor) DoFunc(handler MessageSinkFunc) *Reactor {
	return a.Do(handler)
}

// ContextRunner defines a runner accepts a context
// the runner should be started using go runner.Run(ctx)
type ContextRunner interface {
	Run(context.Context)
}

// ImmediateFuture implements a future with immediate result
type ImmediateFuture struct {
	Error error
}

// Wait implements Future
func (f *ImmediateFuture) Wait() error {
	return f.Error
}
