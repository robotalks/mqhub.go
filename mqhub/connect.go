package mqhub

import (
	"net/url"
	"strings"
)

// Protocol defines the protocol in the URL
// A few examples of URL acceptable:
//   - mqhub+mqtt://server:port/topic
//   - mqhub+mqtt+ws://server:port/topic
//   - mqtt://server:port/topic
//   - mqtt+ws://server:port/topic
const Protocol = "mqhub"

// ConnectorFactory creates a connector from a URL
type ConnectorFactory func(URL url.URL) (Connector, error)

var _connectorFactories = make(map[string]ConnectorFactory)

// RegisterConnectorFactory registers a ConnectorFactory for the specified protocol
func RegisterConnectorFactory(protocol string, factory ConnectorFactory) {
	_connectorFactories[protocol] = factory
}

// NewConnector creates a connector by parsing URL
func NewConnector(URL string) (Connector, error) {
	parsedURL, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}
	return NewConnectorURL(*parsedURL)
}

// NewConnectorURL creates a connector from a parsed URL
func NewConnectorURL(connURL url.URL) (Connector, error) {
	if strings.HasPrefix(connURL.Scheme, Protocol+"+") {
		connURL.Scheme = connURL.Scheme[len(Protocol)+1:]
	}
	prot := connURL.Scheme
	// if protocol is like mqtt+tcp, the portion before + is used to match
	// the protocol name, that allows ConnectorFactory to choose alternative
	// lower level transportation
	if pos := strings.Index(prot, "+"); pos >= 0 {
		prot = prot[:pos]
	}
	factory := _connectorFactories[prot]
	if factory == nil {
		return nil, nil
	}
	return factory(connURL)
}
