package mqhub

import "fmt"

var (
	// ErrNoMessageSink is reported by DataPoint when message sink is not yet
	// connected
	ErrNoMessageSink = fmt.Errorf("message sink unavailable")
)
