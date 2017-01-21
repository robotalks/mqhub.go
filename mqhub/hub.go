package mqhub

import "io"

// Message is the abstraction of data entity passing through the hub
type Message interface {
	// Component is the ID of component emitted this message
	Component() string
	// Endpoint is the name of the endpoint
	Endpoint() string
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

// Watchable is an object which others can watch for changes
type Watchable interface {
	Watch(MessageSink) (Watcher, error)
}

// Watcher is an established watching state
type Watcher interface {
	io.Closer
	Watched() Watchable
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

// Connector defines a general connector to message source/bus
type Connector interface {
	io.Closer
	Watchable
	Connect() Future
	Publish(Component) (Publication, error)
	Describe(componentID string) Descriptor
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

// Descriptor describes a publication
type Descriptor interface {
	Watchable
	ID() string
	SubComponent(id ...string) Descriptor
	Endpoint(name string) EndpointRef
}

// EndpointRef references remote endpoints
type EndpointRef interface {
	Watchable
	MessageSink
}
