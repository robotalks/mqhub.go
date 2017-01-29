package mqtt

import (
	"crypto/tls"
	"net/url"
	"strings"
	"sync"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/robotalks/mqhub.go/mqhub"
	"github.com/robotalks/mqhub.go/utils"
)

// Options defines configuration for connection
type Options struct {
	Servers   []*url.URL
	TLS       tls.Config
	Username  string
	Password  string
	ClientID  string
	Namespace string
}

// NewOptions creates options
func NewOptions() *Options {
	return &Options{}
}

// AddServer adds a server
func (o *Options) AddServer(server string) *Options {
	serverURL, err := url.Parse(server)
	if err != nil {
		panic(err)
	}
	o.Servers = append(o.Servers, serverURL)
	return o
}

// Auth sets username and password
func (o *Options) Auth(username, password string) *Options {
	o.Username = username
	o.Password = password
	return o
}

// SetClientID sets client ID
func (o *Options) SetClientID(id string) *Options {
	o.ClientID = id
	return o
}

func (o *Options) clientOptions() *paho.ClientOptions {
	opts := paho.NewClientOptions()
	opts.Servers = o.Servers
	opts.ClientID = o.ClientID
	opts.Username = o.Username
	opts.Password = o.Password
	opts.ProtocolVersion = 3
	if opts.ClientID == "" {
		opts.ClientID = utils.UniqueID()
	}
	return opts
}

// Connector connects to MQTT
type Connector struct {
	Client paho.Client

	topicPrefix string
	exports     []*Publication
	lock        sync.RWMutex
	handlers    *TopicHandlerMap
}

// NewConnector creates a connector
func NewConnector(options *Options) *Connector {
	if options == nil {
		options = NewOptions()
	}
	conn := &Connector{
		Client:      paho.NewClient(options.clientOptions()),
		topicPrefix: options.Namespace,
		handlers:    NewTopicHandlerMap(),
	}
	if conn.topicPrefix != "" && !strings.HasSuffix(conn.topicPrefix, "/") {
		conn.topicPrefix += "/"
	}
	conn.handlers.prefix = conn.topicPrefix
	return conn
}

// Watch implements Watchable
func (c *Connector) Watch(sink mqhub.MessageSink) (mqhub.Watcher, error) {
	return watchTopic(c, c, "#", sink)
}

// Connect connects to server
func (c *Connector) Connect() mqhub.Future {
	return &Future{token: c.Client.Connect()}
}

// Close implements io.Closer
func (c *Connector) Close() error {
	c.Client.Disconnect(0)
	return nil
}

// Publish implements Publisher
func (c *Connector) Publish(comp mqhub.Component) (mqhub.Publication, error) {
	pub := newPublication(c, comp)
	c.lock.Lock()
	c.exports = append(c.exports, pub)
	c.lock.Unlock()
	if err := pub.export(); err != nil {
		pub.unexport()
		c.removePub(pub)
		return nil, err
	}
	return pub, nil
}

// Describe creates a descriptor
func (c *Connector) Describe(componentID string) mqhub.Descriptor {
	return &Descriptor{
		ComponentID: componentID,
		SubTopic:    componentID,
		conn:        c,
	}
}

func (c *Connector) parseTopic(topic string) (string, string) {
	return ParseTopic(topic, c.topicPrefix)
}

func (c *Connector) newMsg(msg paho.Message) *Message {
	return NewMessage(c.topicPrefix, msg)
}

func (c *Connector) sub(topics []string, handler *HandlerRef) *Future {
	subs := c.handlers.Add(topics, handler)
	subsMap := make(map[string]byte)
	for _, topic := range subs {
		subsMap[c.topicPrefix+topic] = 0
	}
	return &Future{token: c.Client.SubscribeMultiple(subsMap, c.handlers.HandleMessage)}
}

func (c *Connector) unsub(topics []string, handler *HandlerRef) *Future {
	unsubs := c.handlers.Del(topics, handler)
	for i, topic := range unsubs {
		unsubs[i] = c.topicPrefix + topic
	}
	return &Future{token: c.Client.Unsubscribe(unsubs...)}
}

func (c *Connector) pub(topic string, msg mqhub.Message) *Future {
	encoded, err := Encode(msg)
	if err != nil {
		return &Future{err: err}
	}
	return &Future{token: c.Client.Publish(c.topicPrefix+topic, 0, msg.IsState(), encoded)}
}

func (c *Connector) removePub(pub *Publication) {
	c.lock.Lock()
	for i, x := range c.exports {
		if x == pub {
			c.exports = append(c.exports[:i], c.exports[i+1:]...)
			break
		}
	}
	c.lock.Unlock()
}

// ConnectorFactory implements mqhub.ConnectorFactory
func ConnectorFactory(URL url.URL) (mqhub.Connector, error) {
	opts := NewOptions()
	if strings.HasPrefix(URL.Scheme, Protocol+"+") {
		URL.Scheme = URL.Scheme[len(Protocol)+1:]
	} else if URL.Scheme == Protocol {
		URL.Scheme = "tcp"
	}
	opts.Servers = append(opts.Servers, &URL)
	if URL.User != nil {
		opts.Username = URL.User.Username()
		if pwd, exist := URL.User.Password(); exist {
			opts.Password = pwd
		}
	}
	opts.Namespace = strings.Trim(URL.Path, "/")
	for key, vals := range URL.Query() {
		switch key {
		case OptClientID:
			if len(vals) > 0 {
				unescaped, err := url.QueryUnescape(vals[len(vals)-1])
				if err != nil {
					return nil, err
				}
				opts.ClientID = unescaped
			}
		}
	}
	return NewConnector(opts), nil
}

const (
	// Protocol is the name of protocol for connector
	Protocol = "mqtt"

	// OptClientID is the property name in URL query
	OptClientID = "client-id"
)

func init() {
	mqhub.RegisterConnectorFactory(Protocol, ConnectorFactory)
}
