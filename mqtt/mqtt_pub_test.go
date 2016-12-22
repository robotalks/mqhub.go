package mqtt_test

import (
	"log"
	"testing"

	"github.com/robotalks/mqhub.go/mqhub"
	"github.com/robotalks/mqhub.go/mqtt"
	"github.com/robotalks/mqhub.go/utils"
	"github.com/stretchr/testify/assert"
)

type Comp0 struct {
	mqhub.ComponentBase
	state0 mqhub.DataPoint
	state1 mqhub.DataPoint
	actor0 mqhub.Reactor
}

func NewComp0(id string) *Comp0 {
	c := &Comp0{
		state0: mqhub.DataPoint{Name: "state0", Retain: true},
		state1: mqhub.DataPoint{Name: "state1"},
		actor0: mqhub.Reactor{Name: "a"},
	}
	c.SetID(id)
	c.actor0.Handler = mqhub.MessageSinkAs(c.actor)
	return c
}

func (c *Comp0) Endpoints() []mqhub.Endpoint {
	return []mqhub.Endpoint{&c.state0, &c.state1, &c.actor0}
}

func (c *Comp0) actor(state int) {
	c.state0.Update(state)
}

type Pub0 struct {
	Comp0 *Comp0
}

func NewPub0() *Pub0 {
	return &Pub0{Comp0: NewComp0("comp0")}
}

func (p *Pub0) ID() string {
	return "pub0"
}

func (p *Pub0) Components() []mqhub.Component {
	return []mqhub.Component{p.Comp0}
}

func (p *Pub0) Endpoints() []mqhub.Endpoint {
	return nil
}

type pubState struct {
	state     int
	component string
	endpoint  string
}

func makeSinkFunc(t *testing.T, ch chan pubState) mqhub.MessageSink {
	return mqhub.MessageSinkFunc(func(msg mqhub.Message) mqhub.Future {
		var state pubState
		assert.NoError(t, msg.As(&state.state))
		state.component = msg.Component()
		state.endpoint = msg.Endpoint()
		log.Printf("MSG %s: %d", msg.(*mqtt.Message).Raw.Topic(), state.state)
		ch <- state
		return &mqhub.ImmediateFuture{}
	})
}

func TestPublication(t *testing.T) {
	a := assert.New(t)
	prefix := "pub-" + utils.UniqueID()

	host, err := TestEnv.NewConnector(prefix, "publication-pub")
	if !a.NoError(err) || !a.NoError(host.Connect().Wait()) {
		return
	}

	pub0 := NewPub0()

	_, err = host.Publish(pub0)
	if !a.NoError(err) {
		return
	}
	a.NoError(pub0.Comp0.state0.Update(1).Wait())
	a.NoError(pub0.Comp0.state0.Update(2).Wait())
	a.NoError(pub0.Comp0.state1.Update(20).Wait())

	client, err := TestEnv.NewConnector(prefix, "publication-client")
	if !a.NoError(err) || !a.NoError(client.Connect().Wait()) {
		return
	}

	stateCh := make(chan pubState, 1)
	sinkFunc := makeSinkFunc(t, stateCh)

	desc := client.Describe("pub0")
	_, err = desc.Endpoint("comp0", "state0").Watch(sinkFunc)
	a.NoError(err)
	state := <-stateCh
	a.Equal(2, state.state)
	a.Equal("pub0/comp0", state.component)
	a.Equal("state0", state.endpoint)
	actor, err := desc.Endpoint("comp0", "a").Reactor()
	a.NoError(err)
	err = actor.ConsumeMessage(mqhub.MsgFrom(100)).Wait()
	if !a.NoError(err) {
		return
	}
	state = <-stateCh
	a.Equal("pub0/comp0", state.component)
	a.Equal("state0", state.endpoint)
	a.Equal(100, state.state)

	_, err = desc.Endpoint("comp0", "state1").Watch(sinkFunc)
	a.NoError(err)
	a.NoError(pub0.Comp0.state1.Update(30).Wait())
	state = <-stateCh
	a.Equal(30, state.state)
	a.Equal("pub0/comp0", state.component)
	a.Equal("state1", state.endpoint)
}

func TestTopLevelSubscribe(t *testing.T) {
	a := assert.New(t)
	prefix := "toplevel-sub-" + utils.UniqueID()

	host, err := TestEnv.NewConnector(prefix, "subscribe-toplevel")
	if !a.NoError(err) || !a.NoError(host.Connect().Wait()) {
		return
	}
	defer host.Close()

	pub0 := NewPub0()

	_, err = host.Publish(pub0)
	if !a.NoError(err) {
		return
	}

	client, err := TestEnv.NewConnector(prefix, "subscribe-toplevel-client")
	if !a.NoError(err) || !a.NoError(client.Connect().Wait()) {
		return
	}
	defer client.Close()

	stateCh := make(chan pubState, 2)
	var watcher mqhub.Watcher
	watcher, err = client.Watch(makeSinkFunc(t, stateCh))
	a.NoError(err)
	defer watcher.Close()

	desc := client.Describe("pub0")
	actor, err := desc.Endpoint("comp0", "a").Reactor()
	if !a.NoError(err) {
		return
	}
	a.NoError(actor.ConsumeMessage(mqhub.MsgFrom(101)).Wait())
	state := <-stateCh
	a.Equal(101, state.state)
	a.Equal("pub0/comp0", state.component)
	a.Equal("state0", state.endpoint)

	a.NoError(pub0.Comp0.state1.Update(201).Wait())
	state = <-stateCh
	a.Equal(201, state.state)
	a.Equal("pub0/comp0", state.component)
	a.Equal("state1", state.endpoint)
}
