package mqtt

import "github.com/robotalks/mqhub.go/mqhub"

// Descriptor represents advertisements
// and implements mqhub.Descriptor
type Descriptor struct {
	ComponentID string `json:"id"`
	SubTopic    string `json:"topic"`

	conn *Connector
}

// SubComponent implements Descriptor
func (d *Descriptor) SubComponent(id ...string) mqhub.Descriptor {
	if len(id) == 0 {
		return d
	}
	return &Descriptor{
		ComponentID: id[len(id)-1],
		SubTopic:    SubCompTopic(d.SubTopic, id...),
		conn:        d.conn,
	}
}

// Watch implements Descriptor
func (d *Descriptor) Watch(sink mqhub.MessageSink) (mqhub.Watcher, error) {
	return watchTopic(d.conn, d, SubCompTopic(d.SubTopic, "#"), sink)
}

// ID implements Descriptor
func (d *Descriptor) ID() string {
	return d.ComponentID
}

// Endpoint implements Descriptor
func (d *Descriptor) Endpoint(name string) mqhub.EndpointRef {
	return &EndpointRef{
		conn:  d.conn,
		topic: EndpointTopic(d.SubTopic, name),
	}
}

// EndpointRef implements mqhub.EndpointRef
type EndpointRef struct {
	conn  *Connector
	topic string
}

// Watch implements EndpointRef
func (r *EndpointRef) Watch(sink mqhub.MessageSink) (mqhub.Watcher, error) {
	return watchTopic(r.conn, r, r.topic, sink)
}

// ConsumeMessage implements MessageSink
func (r *EndpointRef) ConsumeMessage(msg mqhub.Message) mqhub.Future {
	return r.conn.pub(r.topic, msg)
}
