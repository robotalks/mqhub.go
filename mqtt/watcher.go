package mqtt

import (
	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/robotalks/mqhub.go/mqhub"
)

type topicWatcher struct {
	conn    *Connector
	target  mqhub.Watchable
	topic   string
	sink    mqhub.MessageSink
	handler *HandlerRef
}

func watchTopic(conn *Connector, target mqhub.Watchable,
	topic string, sink mqhub.MessageSink) (*topicWatcher, error) {
	w := &topicWatcher{
		conn:   conn,
		target: target,
		topic:  topic,
		sink:   sink,
	}
	w.handler = MakeHandlerRef(w.recvMessage)
	return w, w.conn.sub([]string{topic}, w.handler).Wait()
}

// Close implements Watcher
func (w *topicWatcher) Close() error {
	w.conn.unsub([]string{w.topic}, w.handler)
	return nil
}

// Watched implements Watcher
func (w *topicWatcher) Watched() mqhub.Watchable {
	return w.target
}

func (w *topicWatcher) recvMessage(_ paho.Client, msg paho.Message) {
	_, endpoint := w.conn.parseTopic(msg.Topic())
	if endpoint != "" {
		w.sink.ConsumeMessage(w.conn.newMsg(msg))
	}
}
