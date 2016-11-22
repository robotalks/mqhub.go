package mqhub

import (
	"context"
)

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
	return p.Sink.ConsumeMessage(MakeMsg(state, p.Retain))
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
