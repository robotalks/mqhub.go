package mqhub

// OriginMsg wraps existing value
type OriginMsg struct {
	ComponentID  string
	EndpointName string
	V            interface{}
	State        bool
}

// Component implements Message
func (m *OriginMsg) Component() string {
	return m.ComponentID
}

// Endpoint implements Message
func (m *OriginMsg) Endpoint() string {
	return m.EndpointName
}

// Value implements Message
func (m *OriginMsg) Value() (interface{}, bool) {
	return m.V, true
}

// IsState implements Message
func (m *OriginMsg) IsState() bool {
	return m.State
}

// As implements Message
func (m *OriginMsg) As(interface{}) error {
	return nil
}

// MakeMsg creates an OriginMsg
func MakeMsg(v interface{}, state bool) *OriginMsg {
	return &OriginMsg{V: v, State: state}
}

// MsgFrom makes a stateless message
func MsgFrom(v interface{}) *OriginMsg {
	return MakeMsg(v, false)
}

// StateFrom makes a stateful message
func StateFrom(v interface{}) *OriginMsg {
	return MakeMsg(v, true)
}

// StreamMessage implements Message with raw bytes
type StreamMessage []byte

// Component implements Message
func (m StreamMessage) Component() string {
	return ""
}

// Endpoint implements Message
func (m StreamMessage) Endpoint() string {
	return ""
}

// Value implements mqhub.Message
func (m StreamMessage) Value() (interface{}, bool) {
	return m, true
}

// IsState implements Message
func (m StreamMessage) IsState() bool {
	return false
}

// As implements Message
func (m StreamMessage) As(out interface{}) error {
	return nil
}

// Payload implements EncodedPayload
func (m StreamMessage) Payload() ([]byte, error) {
	return m, nil
}
