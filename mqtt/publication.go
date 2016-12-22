package mqtt

import (
	"path"
	"regexp"
	"strings"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/robotalks/mqhub.go/mqhub"
)

// DataTopic creates a topic for data points
func DataTopic(topicBase, name string) string {
	return path.Join(topicBase, ":", name)
}

// ActorTopic creates a topic for reactors
func ActorTopic(topicBase, name string) string {
	return path.Join(topicBase, "!", name)
}

// SubCompTopic creates a topic for sub-components
func SubCompTopic(topicBase, id string) string {
	return path.Join(topicBase, id)
}

var topicRelRe = regexp.MustCompile(`^(.+)/([:!])/([^/]+)$`)

// ParseTopicRel parse topic without prefix (start with component ID)
func ParseTopicRel(relativeTopic string) (compID, endpoint, endpointType string) {
	result := topicRelRe.FindAllStringSubmatch(relativeTopic, -1)
	if len(result) > 0 && len(result[0]) >= 4 {
		return result[0][1], result[0][3], result[0][2]
	}
	return relativeTopic, "", ""
}

// ParseTopic parse topic including the prefix
func ParseTopic(topic, prefix string) (compID, endpoint, endpointType string) {
	if !strings.HasPrefix(topic, prefix) {
		return topic, "", ""
	}
	return ParseTopicRel(topic[len(prefix):])
}

// Publication implements mqhub.Publication
type Publication struct {
	conn  *Connector
	comp  mqhub.Component
	topic string
	desc  Descriptor
	emits map[string]*DataEmitter
	sinks map[string]*DataSink
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
		conn:  conn,
		comp:  comp,
		topic: SubCompTopic(conn.topicPrefix, comp.ID()),
		emits: make(map[string]*DataEmitter),
		sinks: make(map[string]*DataSink),
	}
	pub.desc.ComponentID = comp.ID()
	pub.desc.Topic = pub.topic
	pub.populate(pub.topic, comp)
	return pub
}

func (p *Publication) populate(topic string, comp mqhub.Component) {
	endpoints := comp.Endpoints()
	for _, endpoint := range endpoints {
		if datapoint, ok := endpoint.(mqhub.MessageSource); ok {
			endpointTopic := DataTopic(topic, endpoint.ID())
			p.emits[endpointTopic] = &DataEmitter{
				pub:    p,
				topic:  endpointTopic,
				source: datapoint,
			}
		}
		if reactor, ok := endpoint.(mqhub.MessageSink); ok {
			endpointTopic := ActorTopic(topic, endpoint.ID())
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
	topics := make(map[string]byte)
	for topic := range p.sinks {
		topics[topic] = 0
	}
	token := p.conn.Client.SubscribeMultiple(topics, p.handleMessage)
	token.Wait()
	err := token.Error()
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
	p.conn.Client.Unsubscribe(topics...)
}

func (p *Publication) handleMessage(_ paho.Client, msg paho.Message) {
	if sink := p.sinks[msg.Topic()]; sink != nil {
		sink.ConsumeMessage(NewMessage(p.conn.topicPrefix, msg))
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
	encoded, err := Encode(msg)
	if err != nil {
		return &Future{err: err}
	}
	return &Future{token: e.pub.conn.Client.Publish(e.topic, 0, msg.IsState(), encoded)}
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
