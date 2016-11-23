package mqtt

import (
	"encoding/json"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/robotalks/mqhub.go/mqhub"
)

// Message implements Message
type Message struct {
	raw paho.Message
}

// NewMessage wraps mqtt message
func NewMessage(msg paho.Message) *Message {
	return &Message{raw: msg}
}

// Value implements Message
func (m *Message) Value() (interface{}, bool) {
	// this is a wrapper over mqtt message, no value
	return nil, false
}

// IsState implements Message
func (m *Message) IsState() bool {
	return m.raw.Retained()
}

// As implements Message
func (m *Message) As(out interface{}) error {
	if data := m.raw.Payload(); data != nil {
		return json.Unmarshal(data, out)
	}
	return nil
}

// Payload implements EncodedPayload
func (m *Message) Payload() ([]byte, error) {
	return m.raw.Payload(), nil
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
