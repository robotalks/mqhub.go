package mqtt

import (
	"strings"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/robotalks/mqhub.go/mqhub"
)

type prefixWatcher struct {
	client paho.Client
	target mqhub.Watchable
	conn   *Connector
	prefix string
	msgCh  chan paho.Message
}

func watchPrefix(client paho.Client, target mqhub.Watchable,
	prefix string, sink mqhub.MessageSink) (*prefixWatcher, error) {
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	w := &prefixWatcher{
		client: client,
		target: target,
		prefix: prefix,
		msgCh:  make(chan paho.Message),
	}
	return w, w.watch(sink)
}

// Close implements Watcher
func (w *prefixWatcher) Close() error {
	close(w.msgCh)
	return nil
}

// Watched implements Watcher
func (w *prefixWatcher) Watched() mqhub.Watchable {
	return w.target
}

func (w *prefixWatcher) watch(sink mqhub.MessageSink) error {
	token := w.client.Subscribe(w.prefix+"#", 0, w.recvMessage)
	token.Wait()
	if err := token.Error(); err != nil {
		return err
	}
	go w.run(sink)
	return nil
}

func (w *prefixWatcher) recvMessage(_ paho.Client, msg paho.Message) {
	_, _, endpointType := ParseTopic(msg.Topic(), w.prefix)
	if endpointType == ":" {
		w.msgCh <- msg
	}
}

func (w *prefixWatcher) run(sink mqhub.MessageSink) {
	for {
		msg, ok := <-w.msgCh
		if !ok {
			break
		}
		sink.ConsumeMessage(NewMessage(w.prefix, msg))
	}
	w.client.Unsubscribe(w.prefix + "#")
}
