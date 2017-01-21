package mqtt

import (
	"encoding/json"
	"path"
	"strings"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/robotalks/mqhub.go/mqhub"
)

// EndpointTopic creates a topic for endpoints
func EndpointTopic(topicBase, name string) string {
	return path.Join(topicBase, name)
}

// SubCompTopic creates a topic for sub-components
func SubCompTopic(topicBase string, id ...string) string {
	return path.Join(append([]string{topicBase}, id...)...)
}

// ParseTopicRel parse topic without prefix (start with component ID)
func ParseTopicRel(relativeTopic string) (compID, endpoint string) {
	compID, endpoint = path.Split(relativeTopic)
	compID = strings.Trim(compID, "/")
	return
}

// ParseTopic parse topic including the prefix
func ParseTopic(topic, prefix string) (compID, endpoint string) {
	if prefix == "" {
		return ParseTopicRel(topic)
	}
	if !strings.HasPrefix(topic, prefix) {
		return topic, ""
	}
	return ParseTopicRel(topic[len(prefix):])
}

// TokenizeTopic split topic into tokens
func TokenizeTopic(topic string) []string {
	return strings.Split(topic, "/")
}

// TopicFilter defines a parsed topic filter
type TopicFilter struct {
	tokens      []string
	multiLevels bool
}

// NewTopicFilter parses a topic filter in string
func NewTopicFilter(filter string) *TopicFilter {
	f := &TopicFilter{}
	if filter == "#" {
		f.multiLevels = true
		return f
	}

	if strings.HasSuffix(filter, "/#") {
		f.multiLevels = true
		filter = filter[:len(filter)-2]
	}
	f.tokens = TokenizeTopic(filter)
	return f
}

// String returns the filter in string
func (f *TopicFilter) String() string {
	s := strings.Join(f.tokens, "/")
	if f.multiLevels {
		if len(f.tokens) > 0 {
			s += "/#"
		} else {
			s = "#"
		}
	}
	return s
}

// MatchesTokenized indicates f matches the tokenized topic
func (f *TopicFilter) MatchesTokenized(tokens []string) bool {
	if len(f.tokens) > len(tokens) {
		return false
	}
	for i := 0; i < len(f.tokens); i++ {
		if f.tokens[i] != "+" && f.tokens[i] != tokens[i] {
			return false
		}
	}
	if len(f.tokens) < len(tokens) && !f.multiLevels {
		return false
	}
	return true
}

// Matches indicates f matches the topic
func (f *TopicFilter) Matches(topic string) bool {
	return f.MatchesTokenized(strings.Split(topic, "/"))
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
	m.ComponentID, m.EndpointName = ParseTopic(msg.Topic(), prefix)
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
