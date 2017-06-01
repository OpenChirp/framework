// Package service provides the management library for a long running service
package framework

import (
	"errors"
	"log"
	"math/big"

	CRAND "crypto/rand"

	"os"

	"encoding/json"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/openchirp/framework/rest"
)

const (
	deviceUpdatesBuffering = 10
	mqttPersistence        = false // we should never have this enabled
)

/* Options to be filled in by arguments */
var mqttBroker string
var mqttUser string
var mqttPass string
var mqttQos uint

var ErrNotImplemented = errors.New("This method is not implemented yet")

const (
	// DeviceUpdateAdd indicates that a new device linked in this service
	DeviceUpdateTypeAdd = iota
	// DeviceUpdateRem indicates a device has unlinked this service
	DeviceUpdateTypeRem
	// DeviceUpdateUpd indicates that a device this service's config
	DeviceUpdateTypeUpd
)

type DeviceUpdate struct {
	Type int
	ServiceDeviceUpdate
}

// TopicHandler is a function prototype for a subscribed topic callback
type TopicHandler func(service *Service, topic string, payload []byte)

type Service struct {
	id      string
	host    rest.Host
	mqtt    MQTT.Client
	node    rest.ServiceNode
	updates chan DeviceUpdate
	log     *log.Logger
}

// genClientID generates a random client id for mqtt
func (s Service) genClientID() string {
	r, err := CRAND.Int(CRAND.Reader, new(big.Int).SetInt64(100000))
	if err != nil {
		log.Fatal("Couldn't generate a random number for MQTT client ID")
	}
	return s.node.ID + "-" + r.String()
}

// CreateService creates the named service on the framework server
// and returns serviceid upon sucess
// func CreateService(host framework.Host, name string) (string, error) {
// 	host = host // exercise that host variable
// 	name = name // exercise that name variable
// 	return "", ErrNotImplemented
// }

// StartService starts the service maangement layer for service
// with id serviceID
func StartService(host rest.Host, serviceID string) (*Service, error) {
	var err error

	s := new(Service)
	s.host = host
	s.id = serviceID
	s.log = log.New(os.Stderr, "Service:", log.Flags())

	// we should expect mqtt settings to come from framework host
	// for now, we will simply deduce it from framework Host
	// url.Parse(host.)

	// Get Our Service Info
	s.node, err = s.host.RequestServiceInfo(s.id)
	if err != nil {
		return nil, err
	}

	// Connect to MQTT
	/* Setup basic MQTT connection */
	opts := MQTT.NewClientOptions().AddBroker(s.node.Properties["MQTTBroker"])
	opts.SetClientID(s.genClientID())
	opts.SetUsername(s.node.Properties["MQTTUser"])
	opts.SetPassword(s.node.Properties["MQTTPass"])

	/* Create and start a client using the above ClientOptions */
	s.mqtt = MQTT.NewClient(opts)
	if token := s.mqtt.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()

	}

	return s, nil
}

func (s *Service) StartDeviceUpdates() (<-chan DeviceUpdate, error) {
	s.updates = make(chan DeviceUpdate, deviceUpdatesBuffering)
	// Hack until we have one unified topic
	topicAdd := s.node.Pubsub.Topic + "/thing/new"
	topicRem := s.node.Pubsub.Topic + "/thing/remove"
	topicUpd := s.node.Pubsub.Topic + "/thing/update"
	err := s.Subscribe(topicAdd, func(service *Service, topic string, payload []byte) {
		var mqttMsg ServiceUpdatesEncapsulation
		err := json.Unmarshal(payload, &mqttMsg)
		if err != nil {
			s.log.Printf("Failed to unmarshal message on topic %s\n", topic)
			return
		}
		s.updates <- DeviceUpdate{
			Type:                DeviceUpdateTypeAdd,
			ServiceDeviceUpdate: mqttMsg.Thing,
		}
	})
	if err != nil {
		close(s.updates)
		s.updates = nil
	}

	err = s.Subscribe(topicRem, func(service *Service, topic string, payload []byte) {
		var mqttMsg ServiceUpdatesEncapsulation
		err := json.Unmarshal(payload, &mqttMsg)
		if err != nil {
			s.log.Printf("Failed to unmarshal message on topic %s\n", topic)
			return
		}
		s.updates <- DeviceUpdate{
			Type:                DeviceUpdateTypeRem,
			ServiceDeviceUpdate: mqttMsg.Thing,
		}
	})
	if err != nil {
		s.Unsubscribe(topicAdd)
		close(s.updates)
		s.updates = nil
	}

	err = s.Subscribe(topicUpd, func(service *Service, topic string, payload []byte) {
		var mqttMsg ServiceUpdatesEncapsulation
		err := json.Unmarshal(payload, &mqttMsg)
		if err != nil {
			s.log.Printf("Failed to unmarshal message on topic %s\n", topic)
			return
		}
		s.updates <- DeviceUpdate{
			Type:                DeviceUpdateTypeUpd,
			ServiceDeviceUpdate: mqttMsg.Thing,
		}
	})
	if err != nil {
		s.Unsubscribe(topicAdd)
		s.Unsubscribe(topicRem)
		close(s.updates)
		s.updates = nil
	}

	return s.updates, err
}

func (s *Service) StopDeviceUpdates() {
	// Hack until we have one unified topic
	topicAdd := s.node.Pubsub.Topic + "/thing/new"
	topicRem := s.node.Pubsub.Topic + "/thing/remove"
	topicUpd := s.node.Pubsub.Topic + "/thing/update"
	s.Unsubscribe(topicAdd)
	s.Unsubscribe(topicRem)
	s.Unsubscribe(topicUpd)
	close(s.updates)
}

func (s *Service) FetchDeviceConfigs() ([]rest.ServiceDeviceListItem, error) {
	// Get The Current Device Config
	devs, err := s.host.RequestServiceDeviceList(s.id)
	return devs, err
}

// StopService shuts down a started service
func (s *Service) StopService() {
	s.mqtt.Disconnect(0)
}

func (s *Service) Subscribe(topic string, callback TopicHandler) error {
	token := s.mqtt.Subscribe(topic, byte(mqttQos), func(client MQTT.Client, message MQTT.Message) {
		callback(s, message.Topic(), message.Payload())
	})
	token.Wait()
	return token.Error()
}

func (s *Service) Unsubscribe(topic string) error {
	token := s.mqtt.Unsubscribe(topic)
	token.Wait()
	return token.Error()
}

func (s *Service) Publish(topic string, payload []byte) error {
	token := s.mqtt.Publish(topic, byte(mqttQos), mqttPersistence, payload)
	token.Wait()
	return token.Error()
}

// GetProperties returns the full properties key/value mapping
func (s *Service) GetProperties() map[string]string {
	return s.node.Properties
}

// GetProperty fetches the service property associated with key. If it does
// not exist the blank string is returned.
func (s *Service) GetProperty(key string) string {
	value, ok := s.node.Properties[key]
	if ok {
		return value
	}
	return ""
}

func (s *Service) GetMQTTClient() *MQTT.Client {
	return &s.mqtt
}
