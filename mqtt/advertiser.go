package mqtt

import (
	"sync"

	"github.com/robotalks/mqhub.go/mqhub"
)

// Advertiser implements mqhub.Advertiser
type Advertiser struct {
	conn  *Connector
	topic string
	ads   []*Advertisement
	lock  sync.RWMutex
}

// Advertise implements Advertiser
func (a *Advertiser) Advertise(pub mqhub.Publication) (mqhub.Advertisement, error) {
	ad, inserted := a.insert(pub.(*Publication))
	if inserted {
		return ad, a.refresh()
	}
	return ad, nil
}

func newAdvertiser(conn *Connector, topic string) *Advertiser {
	ad := &Advertiser{
		conn:  conn,
		topic: topic,
	}
	return ad
}

func (a *Advertiser) bind() error {
	// TODO
	return nil
}

func (a *Advertiser) refresh() error {
	var descs []Descriptor
	a.lock.RLock()
	for _, ad := range a.ads {
		descs = append(descs, ad.pub.desc)
	}
	a.lock.RUnlock()
	encoded, err := Encode(mqhub.MakeMsg(descs, true))
	if err == nil {
		token := a.conn.Client.Publish(a.topic, 0, true, encoded)
		token.Wait()
		err = token.Error()
	}
	return err
}

func (a *Advertiser) insert(pub *Publication) (*Advertisement, bool) {
	a.lock.Lock()
	defer a.lock.Unlock()
	for _, ad := range a.ads {
		if ad.pub == pub {
			// already exist
			return ad, false
		}
	}
	ad := &Advertisement{
		advertiser: a,
		pub:        pub,
	}
	a.ads = append(a.ads, ad)
	return ad, true
}

func (a *Advertiser) remove(adv *Advertisement) bool {
	a.lock.Lock()
	defer a.lock.Unlock()
	for i, ad := range a.ads {
		if ad == adv {
			a.ads = append(a.ads[:i], a.ads[i+1:]...)
			return true
		}
	}
	return false
}

// Advertisement implements mqhub.Advertisement
type Advertisement struct {
	advertiser *Advertiser
	pub        *Publication
}

// Publication implements Advertisement
func (a *Advertisement) Publication() mqhub.Publication {
	return a.pub
}

// Close implements Advertisement
func (a *Advertisement) Close() error {
	if a.advertiser.remove(a) {
		return a.advertiser.refresh()
	}
	return nil
}
