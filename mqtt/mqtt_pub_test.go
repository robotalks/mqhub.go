package mqtt_test

import (
	"testing"

	"github.com/robotalks/mqhub.go/mqhub"
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

func TestPublication(t *testing.T) {
	a := assert.New(t)
	pub0 := NewPub0()
	host := TestEnv.NewConnector("publication-pub")
	if !a.NoError(host.Connect().Wait()) {
		return
	}

	_, err := host.Publish(pub0)
	if !a.NoError(err) {
		return
	}
	a.NoError(pub0.Comp0.state0.Update(1).Wait())
	a.NoError(pub0.Comp0.state0.Update(2).Wait())
	a.NoError(pub0.Comp0.state1.Update(20).Wait())

	client := TestEnv.NewConnector("publication-client")
	if !a.NoError(client.Connect().Wait()) {
		return
	}

	stateCh := make(chan int, 1)
	sinkFunc := mqhub.MessageSinkFunc(func(msg mqhub.Message) mqhub.Future {
		var state int
		a.NoError(msg.As(&state))
		stateCh <- state
		return &mqhub.ImmediateFuture{}
	})

	desc := client.Describe("pub0")
	_, err = desc.Endpoint("comp0", "state0").Watch(sinkFunc)
	a.NoError(err)
	a.Equal(2, <-stateCh)
	actor, err := desc.Endpoint("comp0", "a").Reactor()
	a.NoError(err)
	err = actor.ConsumeMessage(mqhub.MsgFrom(100)).Wait()
	if !a.NoError(err) {
		return
	}
	a.Equal(100, <-stateCh)

	_, err = desc.Endpoint("comp0", "state1").Watch(sinkFunc)
	a.NoError(err)
	a.NoError(pub0.Comp0.state1.Update(30).Wait())
	a.Equal(30, <-stateCh)
}
