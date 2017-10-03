package pubsub

import (
	"github.com/sirupsen/logrus"
)

// Bridge holds configuration for a PubSub bridge
type Bridge struct {
	pubsuba, pubsubb PubSub
	devicelinks      map[string]links
	log              *logrus.Logger
}

// The typical use case is to only append or overwrite a callback
type links struct {
	// subsubA --> subsubB
	fwd []string
	// subsubB --> subsubA
	rev []string
}

func isIn(arr []string, str string) bool {
	for _, s := range arr {
		if s == str {
			return true
		}
	}
	return false
}

// NewBridge instantiates a PubSub bridge that allows you to map topics from
// one pubsub interface to another and the reverse.
// The log is used to declare errors when publishing asynchronously.
func NewBridge(pubsuba, pubsubb PubSub, log *logrus.Logger) *Bridge {
	b := new(Bridge)
	b.pubsuba = pubsuba
	b.pubsubb = pubsubb
	b.devicelinks = make(map[string]links)
	b.log = log
	// disable logging
	if log == nil {
		b.log = logrus.New()
		b.log.SetLevel(0)
	}
	return b
}

func (b *Bridge) IsDeviceLinked(deviceid string) bool {
	_, ok := b.devicelinks[deviceid]
	return ok
}

func (b *Bridge) IsLinkFwd(deviceid, topica string) bool {
	if ls, ok := b.devicelinks[deviceid]; ok {
		for _, l := range ls.fwd {
			if l == topica {
				return true
			}
		}
	}
	return false
}

func (b *Bridge) IsLinkRev(deviceid, topicb string) bool {
	if ls, ok := b.devicelinks[deviceid]; ok {
		for _, l := range ls.rev {
			if l == topicb {
				return true
			}
		}
	}
	return false
}

func (b *Bridge) AddLinkFwd(deviceid, topica string, topicb ...string) error {
	ls, ok := b.devicelinks[deviceid]
	if !ok {
		ls = links{make([]string, 0), make([]string, 0)}
	}

	// Mark down our link
	if !isIn(ls.fwd, topica) {
		ls.fwd = append(ls.fwd, topica)
	}

	// Subscribe
	err := b.pubsuba.Subscribe(topica, func(topic string, payload []byte) {
		for _, tb := range topicb {
			b.log.Debugf("Received on %s and publishing to %s", topic, tb)
			if err := b.pubsubb.Publish(tb, payload); err != nil {
				b.log.Errorf("Failed to publish to %s: %v", tb, err)
			}
		}

	})
	if err != nil {
		return err
	}

	// Commit changes
	b.devicelinks[deviceid] = ls
	return nil
}

func (b *Bridge) AddFwd(deviceid, topica string, callback func(pubsubb PubSub, topica string, payload []byte) error) error {
	ls, ok := b.devicelinks[deviceid]
	if !ok {
		ls = links{make([]string, 0), make([]string, 0)}
	}

	// Mark down our link
	if !isIn(ls.fwd, topica) {
		ls.fwd = append(ls.fwd, topica)
	}

	err := b.pubsuba.Subscribe(topica, func(topic string, payload []byte) {
		logitem := b.log.WithField("deviceid", deviceid).WithField("topica", topic)
		logitem.Debugf("Running custom callback on received payload")
		if err := callback(b.pubsubb, topic, payload); err != nil {
			logitem.Errorf("Callback reported %v", err)
		}
	})
	if err != nil {
		return err
	}

	// Commit changes
	b.devicelinks[deviceid] = ls
	return nil
}

func (b *Bridge) AddLinkRev(deviceid, topicb string, topica ...string) error {
	ls, ok := b.devicelinks[deviceid]
	if !ok {
		ls = links{make([]string, 0), make([]string, 0)}
	}

	// Mark down our link
	if !isIn(ls.rev, topicb) {
		ls.rev = append(ls.rev, topicb)
	}

	// Subscribe
	err := b.pubsubb.Subscribe(topicb, func(topic string, payload []byte) {
		for _, ta := range topica {
			b.log.Debugf("Received on %s and publishing to %v", topic, ta)
			if err := b.pubsuba.Publish(ta, payload); err != nil {
				b.log.Errorf("Failed to publish to %s: %v", ta, err)
			}
		}

	})
	if err != nil {
		return err
	}

	// Commit changes
	b.devicelinks[deviceid] = ls
	return nil
}

func (b *Bridge) AddRev(deviceid, topicb string, callback func(pubsuba PubSub, topicb string, payload []byte) error) error {
	ls, ok := b.devicelinks[deviceid]
	if !ok {
		ls = links{make([]string, 0), make([]string, 0)}
	}

	// Mark down our link
	if !isIn(ls.rev, topicb) {
		ls.rev = append(ls.rev, topicb)
	}

	err := b.pubsubb.Subscribe(topicb, func(topic string, payload []byte) {
		logitem := b.log.WithField("deviceid", deviceid).WithField("topicb", topic)
		logitem.Debugf("Running custom callback on received payload")
		if err := callback(b.pubsubb, topic, payload); err != nil {
			logitem.Errorf("Callback reported %v", err)
		}
	})
	if err != nil {
		return err
	}

	// Commit changes
	b.devicelinks[deviceid] = ls
	return nil
}

func (b *Bridge) RemoveLinksAll(deviceid string) error {
	var err error
	if ls, ok := b.devicelinks[deviceid]; ok {
		if len(ls.fwd) > 0 {
			e := b.pubsuba.Unsubscribe(ls.fwd...)
			// save and return only first error
			if e != nil && err == nil {
				err = e
			}
		}
		if len(ls.rev) > 0 {
			e := b.pubsubb.Unsubscribe(ls.rev...)
			// save and return only first error
			if e != nil && err == nil {
				err = e
			}
		}
		ls.fwd = nil
		ls.rev = nil
		delete(b.devicelinks, deviceid)
	}
	return err
}
