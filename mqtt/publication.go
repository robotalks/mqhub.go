package mqtt

import (
	"path"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/robotalks/mqhub.go/mqhub"
)

// Publication implements mqhub.Publication
type Publication struct {
	conn    *Connector
	comp    mqhub.Component
	desc    Descriptor
	emits   map[string]*DataEmitter
	sinks   map[string]*DataSink
	handler *HandlerRef
}

// Component implements Publication
func (p *Publication) Component() mqhub.Component {
	return p.comp
}

// Close implements Publication
func (p *Publication) Close() error {
	p.unexport()
	p.conn.removePub(p)
	return nil
}

func newPublication(conn *Connector, comp mqhub.Component) *Publication {
	pub := &Publication{
		conn: conn,
		comp: comp,
		desc: Descriptor{
			ComponentID: comp.ID(),
			SubTopic:    comp.ID(),
			conn:        conn,
		},
		emits: make(map[string]*DataEmitter),
		sinks: make(map[string]*DataSink),
	}
	pub.handler = MakeHandlerRef(pub.handleMessage)
	pub.populate(pub.desc.SubTopic, comp)
	return pub
}

func (p *Publication) populate(topic string, comp mqhub.Component) {
	endpoints := comp.Endpoints()
	for _, endpoint := range endpoints {
		if datapoint, ok := endpoint.(mqhub.MessageSource); ok {
			endpointTopic := EndpointTopic(topic, endpoint.ID())
			p.emits[endpointTopic] = &DataEmitter{
				pub:    p,
				topic:  endpointTopic,
				source: datapoint,
			}
		}
		if reactor, ok := endpoint.(mqhub.MessageSink); ok {
			endpointTopic := EndpointTopic(topic, endpoint.ID())
			p.sinks[endpointTopic] = &DataSink{
				pub:   p,
				topic: endpointTopic,
				sink:  reactor,
			}
		}
	}
	if composite, ok := comp.(mqhub.Composite); ok {
		components := composite.Components()
		for _, c := range components {
			p.populate(path.Join(topic, c.ID()), c)
		}
	}
}

func (p *Publication) export() error {
	topics := make([]string, 0, len(p.sinks))
	for topic := range p.sinks {
		topics = append(topics, topic)
	}
	err := p.conn.sub(topics, p.handler).Wait()
	if err == nil {
		for _, emit := range p.emits {
			emit.bind()
		}
	}
	return err
}

func (p *Publication) unexport() {
	for _, emit := range p.emits {
		emit.unbind()
	}
	topics := make([]string, 0, len(p.sinks))
	p.conn.unsub(topics, p.handler)
}

func (p *Publication) handleMessage(_ paho.Client, msg paho.Message) {
	compID, endpoint := p.conn.parseTopic(msg.Topic())
	if endpoint != "" {
		if sink := p.sinks[EndpointTopic(compID, endpoint)]; sink != nil {
			sink.ConsumeMessage(p.conn.newMsg(msg))
		}
	}
}

// DataEmitter is a consumer which publish the data to hub
type DataEmitter struct {
	pub    *Publication
	topic  string
	source mqhub.MessageSource
}

// ConsumeMessage emits the message
func (e *DataEmitter) ConsumeMessage(msg mqhub.Message) mqhub.Future {
	return e.pub.conn.pub(e.topic, msg)
}

func (e *DataEmitter) bind() {
	e.source.SinkMessage(e)
}

func (e *DataEmitter) unbind() {
	e.source.SinkMessage(nil)
}

// DataSink is a consumer receives messages from hub
type DataSink struct {
	pub   *Publication
	topic string
	sink  mqhub.MessageSink
}

// ConsumeMessage implements MessageSink
func (s *DataSink) ConsumeMessage(msg mqhub.Message) mqhub.Future {
	return s.sink.ConsumeMessage(msg)
}
