package mqtt

import (
	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/robotalks/mqhub.go/mqhub"
)

// Descriptor represents advertisements
// and implements mqhub.Descriptor
type Descriptor struct {
	ComponentID string `json:"id"`
	Topic       string `json:"topic"`

	conn *Connector
}

// ID implements Descriptor
func (d *Descriptor) ID() string {
	return d.ComponentID
}

// Endpoint implements Descriptor
func (d *Descriptor) Endpoint(subPath string, endpoints ...string) mqhub.EndpointRef {
	return &EndpointRef{
		conn:      d.conn,
		topicBase: SubCompTopic(d.Topic, subPath),
		endpoints: endpoints,
	}
}

// EndpointRef implements mqhub.EndpointRef
type EndpointRef struct {
	conn      *Connector
	topicBase string
	endpoints []string
}

// Watch implements EndpointRef
func (r *EndpointRef) Watch(sink mqhub.MessageSink) (mqhub.EndpointWatcher, error) {
	w := &DataPointWatcher{
		conn:      r.conn,
		subTopics: make(map[string]byte),
		msgCh:     make(chan paho.Message),
	}
	for _, name := range r.endpoints {
		topic := DataTopic(r.topicBase, name)
		if _, exist := w.subTopics[topic]; !exist {
			w.subTopics[topic] = 0
		}
	}
	token := r.conn.Client.SubscribeMultiple(w.subTopics, w.recvMessage)
	token.Wait()
	err := token.Error()
	if err == nil {
		go w.run(sink)
	}
	return w, err
}

// Reactor implements EndpointRef
func (r *EndpointRef) Reactor() (mqhub.MessageSink, error) {
	if len(r.endpoints) != 1 {
		panic("exactly one endpoint allowed to connect a reactor")
	}
	return &ReactorConn{conn: r.conn, topic: ActorTopic(r.topicBase, r.endpoints[0])}, nil
}

// DataPointWatcher implements EndpointWatcher
type DataPointWatcher struct {
	conn      *Connector
	subTopics map[string]byte
	msgCh     chan paho.Message
}

// Close implements EndpointWatcher
func (w *DataPointWatcher) Close() error {
	close(w.msgCh)
	return nil
}

func (w *DataPointWatcher) run(sink mqhub.MessageSink) {
	for {
		msg, ok := <-w.msgCh
		if !ok {
			break
		}
		sink.ConsumeMessage(NewMessage(msg))
	}
	topics := make([]string, 0, len(w.subTopics))
	for topic := range w.subTopics {
		topics = append(topics, topic)
	}
	w.conn.Client.Unsubscribe(topics...)
}

func (w *DataPointWatcher) recvMessage(_ paho.Client, msg paho.Message) {
	w.msgCh <- msg
}

// ReactorConn implement ReactorConnection
type ReactorConn struct {
	conn  *Connector
	topic string
}

// Close implements ReactorConnection
func (r *ReactorConn) Close() error {
	return nil
}

// ConsumeMessage implements ReactorConnection
func (r *ReactorConn) ConsumeMessage(msg mqhub.Message) mqhub.Future {
	encoded, err := Encode(msg)
	if err != nil {
		return &Future{err: err}
	}
	token := r.conn.Client.Publish(r.topic, 0, false, encoded)
	return &Future{token: token}
}
