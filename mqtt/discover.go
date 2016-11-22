package mqtt

import (
	"context"

	"github.com/robotalks/mqhub.go/mqhub"
)

// Discoverer implements Discoverer
type Discoverer struct {
	conn  *Connector
	topic string
}

// Discover implements Discoverer
func (d *Discoverer) Discover(ctx context.Context) ([]mqhub.Descriptor, error) {
	// TODO
	return nil, nil
}
