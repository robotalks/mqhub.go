package mqhub

import (
	"context"
	"io"
)

// Message is the abstraction of data entity passing through the hub
type Message interface {
	// Value gets the original value which is not encoded
	Value() (interface{}, bool)
	// IsState indicate this is a datapoint as state
	IsState() bool
	// As decodes the message as the specified type
	As(interface{}) error
}

// EncodedPayload represent an already encoded message
type EncodedPayload interface {
	Payload() ([]byte, error)
}

// Future represent an async operation
type Future interface {
	Wait() error
}

// MessageSink defines the consumer of a message
type MessageSink interface {
	ConsumeMessage(Message) Future
}

// MessageSource emits messages
type MessageSource interface {
	SinkMessage(MessageSink)
}

// Identity represents an object with ID
type Identity interface {
	ID() string
}

// Component is the unit of an object which can be exposed to hub
type Component interface {
	Identity
	// Endpoints enumerates the endpoints this component exposes
	Endpoints() []Endpoint
}

// Endpoint defines the endpoints exposed from a component
// an endpoint can be either a datapoint or a reactor
type Endpoint interface {
	Identity
}

// Composite is a collection of components
type Composite interface {
	Components() []Component
}

// Composer defines composition operations
type Composer interface {
	AddComponent(Component)
}

// Publisher exposes components to hub
type Publisher interface {
	Publish(Component) (Publication, error)
}

// Publication is the expose of the component
type Publication interface {
	io.Closer
	Component() Component
}

// Advertiser makes publications discoverable
type Advertiser interface {
	Advertise(Publication) (Advertisement, error)
}

// Advertisement represents advertised publication
type Advertisement interface {
	io.Closer
	Publication() Publication
}

// Discoverer discovers advertised publications
type Discoverer interface {
	Discover(context.Context) ([]Descriptor, error)
}

// Descriptor describes a publication
type Descriptor interface {
	ID() string
	Endpoint(subPath string, endpoints ...string) EndpointRef
}

// EndpointRef references remote endpoints
type EndpointRef interface {
	// Watch watches data points
	Watch(MessageSink) (EndpointWatcher, error)
	// Reactor connects to a reactor
	Reactor() (MessageSink, error)
}

// EndpointWatcher references watched endpoints
type EndpointWatcher interface {
	io.Closer
}

// ReactorConnection represents the connection to a reactor
type ReactorConnection interface {
	EndpointWatcher
	MessageSink
}
