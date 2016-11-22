package mqhub

// IdentityImpl implements Identity
type IdentityImpl struct {
	id string
}

// ID implements Identity
func (i *IdentityImpl) ID() string {
	return i.id
}

// SetID specifies the ID
func (i *IdentityImpl) SetID(id string) {
	i.id = id
}

// ComponentBase implements Component
type ComponentBase struct {
	IdentityImpl
}

// CompositeBase implements Composite + Composer
type CompositeBase struct {
	components []Component
}

// Components implements Composite
func (c *CompositeBase) Components() []Component {
	return c.components
}

// AddComponent implements Composer
func (c *CompositeBase) AddComponent(comp Component) {
	c.components = append(c.components, comp)
}

// ChanMsgSink is a MessageSink and emits message to a chan
type ChanMsgSink struct {
	C chan Message
}

// NewChanMsgSink creates a ChanMsgSink
func NewChanMsgSink() *ChanMsgSink {
	return &ChanMsgSink{C: make(chan Message)}
}

// ConsumeMessage implements MessageSink
func (c *ChanMsgSink) ConsumeMessage(msg Message) Future {
	c.C <- msg
	return &ImmediateFuture{}
}
