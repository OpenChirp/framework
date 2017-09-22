// Package pubsub hold the interfaces and utilities to work with the PubSub
// side of the OpenChirp framework
package pubsub

// PubSub is the most basic PubSub interface
type PubSub interface {
	Subscribe(topic string, callback func(topic string, payload []byte)) error
	Unsubscribe(topics ...string) error
	Publish(topic string, payload interface{}) error
}
