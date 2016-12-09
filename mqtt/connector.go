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
	advertisers map[string]*Advertiser
	advLock     sync.Mutex
}

// NewConnector creates a connector
func NewConnector(options *Options) *Connector {
	if options == nil {
		options = NewOptions()
	}
	conn := &Connector{
		Client:      paho.NewClient(options.clientOptions()),
		topicPrefix: options.Namespace,
		advertisers: make(map[string]*Advertiser),
	}
	if conn.topicPrefix != "" && !strings.HasSuffix(conn.topicPrefix, "/") {
		conn.topicPrefix += "/"
	}
	return conn
}

// Connect connects to server
func (c *Connector) Connect() mqhub.Future {
	return &Future{token: c.Client.Connect()}
}

// Publish implements Publisher
func (c *Connector) Publish(comp mqhub.Component) (*Publication, error) {
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
func (c *Connector) Describe(componentID string) *Descriptor {
	return &Descriptor{
		ComponentID: componentID,
		Topic:       c.topicPrefix + componentID,
		conn:        c,
	}
}

// Advertiser creates an advertiser on specified topic
func (c *Connector) Advertiser(topic string) (*Advertiser, error) {
	c.advLock.Lock()
	defer c.advLock.Unlock()
	ad := c.advertisers[topic]
	if ad == nil {
		ad = newAdvertiser(c, topic)
		if err := ad.bind(); err != nil {
			return nil, err
		}
		c.advertisers[topic] = ad
	}
	return ad, nil
}

// Discover returns a discoverer associated with a topic
func (c *Connector) Discover(topic string) *Discoverer {
	return &Discoverer{conn: c, topic: topic}
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
