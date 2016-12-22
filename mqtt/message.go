package mqtt

import (
	"encoding/json"
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

// Message implements Message
type Message struct {
	ComponentID  string
	EndpointName string
	Raw          paho.Message
}

// NewMessage wraps mqtt message
func NewMessage(prefix string, msg paho.Message) *Message {
	m := &Message{Raw: msg}
	compID, endpointName, endpointType := ParseTopic(msg.Topic(), prefix)
	if endpointType != "" {
		m.ComponentID = compID
		m.EndpointName = endpointName
	}
	return m
}

// Component implements Message
func (m *Message) Component() string {
	return m.ComponentID
}

// Endpoint implements Message
func (m *Message) Endpoint() string {
	return m.EndpointName
}

// Value implements Message
func (m *Message) Value() (interface{}, bool) {
	// this is a wrapper over mqtt message, no value
	return nil, false
}

// IsState implements Message
func (m *Message) IsState() bool {
	return m.Raw.Retained()
}

// As implements Message
func (m *Message) As(out interface{}) error {
	if data := m.Raw.Payload(); data != nil {
		return json.Unmarshal(data, out)
	}
	return nil
}

// Payload implements EncodedPayload
func (m *Message) Payload() ([]byte, error) {
	return m.Raw.Payload(), nil
}

// Encode encodes original message into bytes
func Encode(msg mqhub.Message) ([]byte, error) {
	if p, ok := msg.(mqhub.EncodedPayload); ok {
		return p.Payload()
	}
	if v, ok := msg.Value(); ok {
		return json.Marshal(v)
	}
	return nil, nil
}

// Future implements mqhub.Future
type Future struct {
	err   error
	token paho.Token
}

// Wait implements Future
func (f *Future) Wait() error {
	if f.token != nil {
		f.token.Wait()
		f.err = f.token.Error()
		f.token = nil
	}
	return f.err
}
