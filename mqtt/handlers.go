package mqtt

import (
	"strings"
	"sync"

	paho "github.com/eclipse/paho.mqtt.golang"
)

// HandlerRef wraps over a handler
type HandlerRef struct {
	Handler paho.MessageHandler
}

// MakeHandlerRef builds a HandlerRef
func MakeHandlerRef(handler paho.MessageHandler) *HandlerRef {
	return &HandlerRef{Handler: handler}
}

// TopicHandlerMap maps topic filter to handlers
type TopicHandlerMap struct {
	prefix string
	topics map[string]*handlerList
	lock   sync.RWMutex
}

type handlerList struct {
	filter   *TopicFilter
	handlers []*HandlerRef
}

// NewTopicHandlerMap creates a new TopicHandlerMap
func NewTopicHandlerMap() *TopicHandlerMap {
	return &TopicHandlerMap{
		topics: make(map[string]*handlerList),
	}
}

// Add inserts filters and corresponding handler
// the returns subs require a SUBSCRIBE
func (m *TopicHandlerMap) Add(filters []string, handler *HandlerRef) (subs []string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	for _, filter := range filters {
		handlers := m.topics[filter]
		if handlers == nil {
			handlers = &handlerList{filter: NewTopicFilter(filter)}
			m.topics[filter] = handlers
			subs = append(subs, filter)
		}
		handlers.handlers = append(handlers.handlers, handler)
	}
	return
}

// Del removes filters/handler, the returned topics require UNSUBSCRIBE
func (m *TopicHandlerMap) Del(filters []string, handler *HandlerRef) (unsubs []string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	for _, filter := range filters {
		handlers := m.topics[filter]
		if handlers != nil {
			for i, h := range handlers.handlers {
				if h == handler {
					handlers.handlers = append(handlers.handlers[:i], handlers.handlers[i+1:]...)
					if len(handlers.handlers) == 0 {
						unsubs = append(unsubs, filter)
						delete(m.topics, filter)
					}
					break
				}
			}
		}
	}
	return
}

// HandleMessage implements paho.MessageHandler
func (m *TopicHandlerMap) HandleMessage(client paho.Client, msg paho.Message) {
	topic := msg.Topic()
	if m.prefix != "" && !strings.HasPrefix(topic, m.prefix) {
		return
	}
	tokens := TokenizeTopic(topic[len(m.prefix):])
	var handlers []*HandlerRef
	m.lock.RLock()
	for _, list := range m.topics {
		if list.filter.MatchesTokenized(tokens) {
			handlers = append(handlers, list.handlers...)
		}
	}
	m.lock.RUnlock()
	for _, handler := range handlers {
		handler.Handler(client, msg)
	}
}
