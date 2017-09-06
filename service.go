// Package service provides the management library for a long running service
// It should be noted that updates
package framework

import (
	"errors"
	"fmt"
	"log"

	"os"

	"encoding/json"

	"github.com/openchirp/framework/rest"
)

const (
	eventsSubTopic         = "/thing/events"
	deviceStatusSubTopic   = "/thing/status"
	statusSubTopic         = "/status"
	deviceUpdatesBuffering = 10
	mqttPersistence        = false // we should never have this enabled
)

/* Options to be filled in by arguments */
var mqttBroker string
var mqttUser string
var mqttPass string
var mqttQos uint

var ErrMarshalStatusMessage = errors.New("Failed to marshall status message into JSON")
var ErrMarshalDeviceStatusMessage = errors.New("Failed to marshall device status message into JSON")
var ErrNotImplemented = errors.New("This method is not implemented yet")

/*
News Updates Look Like The Following:
openchirp/services/592880c57d6ec25f901d9668/thing/events:
{
	"action":"new",
	"thing":{
		"type":"device",
		"id":"5930aaf27d6ec25f901d96da",
		"config":[
			{"key":"rxconfig","value":"[]"},
			{"key":"txconfig","value":"[]"}]
	}
}
*/

type ServiceUpdatesEncapsulation struct {
	Action string                     `json:"action"`
	Device rest.ServiceDeviceListItem `json:"thing"`
}

type ServiceStatus struct {
	Message string `json:"message"`
}

type ServiceDeviceStatus struct {
	Device struct {
		Id      string `json:"id"`
		Message string `json:"message"`
	} `json:"thing"`
}

const (
	// DeviceUpdateAdd indicates that a new device linked in this service
	DeviceUpdateTypeAdd = iota
	// DeviceUpdateRem indicates a device has unlinked this service
	DeviceUpdateTypeRem
	// DeviceUpdateUpd indicates that a device this service's config
	DeviceUpdateTypeUpd
)

// DeviceUpdate represents a pending service config change for a device
type DeviceUpdate struct {
	Type   int
	Id     string
	Config map[string]string
}

// ServiceTopicHandler is a function prototype for a subscribed topic callback
type ServiceTopicHandler func(client *ServiceClient, topic string, payload []byte)

// ServiceClient hold a single ses.Publish(s.)rvice context
type ServiceClient struct {
	Client
	node         rest.ServiceNode
	updatesQueue chan DeviceUpdate
	updates      chan DeviceUpdate
	log          *log.Logger
}

// StartServiceClient starts the service management layer
func StartServiceClient(frameworkuri, brokeruri, id, token string) (*ServiceClient, error) {
	var err error

	c := new(ServiceClient)

	// Start Client
	err = c.startClient(frameworkuri, brokeruri, id, token)
	if err != nil {
		return nil, err
	}

	// Get Our Service Info
	c.node, err = c.host.RequestServiceInfo(c.id)
	if err != nil {
		return nil, err
	}

	c.log = log.New(os.Stderr, "Service:", log.Flags())

	return c, nil
}

// StopClient shuts down a started service
func (c *ServiceClient) StopClient() {
	c.stopClient()
}

// SetStatus publishes the service status message
func (c *ServiceClient) SetStatus(msgs ...interface{}) error {
	var statusmsg ServiceStatus
	statusmsg.Message = fmt.Sprint(msgs...)
	payload, err := json.Marshal(&statusmsg)
	if err != nil {
		return ErrMarshalStatusMessage
	}
	return c.Publish(c.node.Pubsub.Topic+statusSubTopic, payload)
}

// SetDeviceStatus publishes a device's linked service status message
func (c *ServiceClient) SetDeviceStatus(id string, msgs ...interface{}) error {
	var statusmsg ServiceDeviceStatus
	statusmsg.Device.Id = id
	statusmsg.Device.Message = fmt.Sprint(msgs...)
	payload, err := json.Marshal(&statusmsg)
	if err != nil {
		return ErrMarshalDeviceStatusMessage
	}
	return c.Publish(c.node.Pubsub.Topic+deviceStatusSubTopic, payload)
}

// StartDeviceUpdates subscribes to the live mqtt service news topic and opens
// a channel to read the updates from.
// TODO: Services need updates to come from one topic to remove race condition
func (c *ServiceClient) StartDeviceUpdates() (<-chan DeviceUpdate, error) {

	//FIXME: Scheduling between add, remove, and update topic notifications is
	//       inherently a race condition. Please serialize the updates on
	//       the server end into one topic!

	/* Setup MQTT based device updates to feed updatesQueue */
	c.updatesQueue = make(chan DeviceUpdate, deviceUpdatesBuffering)
	topicEvents := c.node.Pubsub.Topic + eventsSubTopic
	err := c.Subscribe(topicEvents, func(service *ServiceClient, topic string, payload []byte) {
		// action: new, update, delete
		var mqttMsg ServiceUpdatesEncapsulation
		var devUpdate DeviceUpdate

		err := json.Unmarshal(payload, &mqttMsg)
		if err != nil {
			c.log.Printf("Failed to unmarshal message on topic %s\n", topic)
			return
		}

		switch mqttMsg.Action {
		case "new":
			devUpdate.Type = DeviceUpdateTypeAdd
		case "update":
			devUpdate.Type = DeviceUpdateTypeUpd
		case "delete":
			devUpdate.Type = DeviceUpdateTypeRem
		}
		devUpdate.Id = mqttMsg.Device.Id
		devUpdate.Config = mqttMsg.Device.GetConfigMap()

		c.updatesQueue <- devUpdate
	})
	if err != nil {
		close(c.updatesQueue)
		return nil, err
	}

	/* Preload device updates from REST request */
	deviceConfigs, err := c.host.RequestServiceDeviceList(c.id)
	if err != nil {
		c.Unsubscribe(topicEvents)
		close(c.updatesQueue)
		c.updatesQueue = nil // make it clear that nobody can reach the chan
		return nil, err
	}
	c.updates = make(chan DeviceUpdate, len(deviceConfigs))
	for _, devConfig := range deviceConfigs {
		c.updates <- DeviceUpdate{
			Type:   DeviceUpdateTypeAdd,
			Id:     devConfig.Id,
			Config: devConfig.GetConfigMap(),
		}
	}

	/* Connect updatesQueue channel to updates channel*/
	go func() {
		for update := range c.updatesQueue {
			c.updates <- update
		}
		close(c.updates)
	}()

	return c.updates, err
}

// StopDeviceUpdates unsubscribes from service news topic and closes the
// news channel
func (c *ServiceClient) StopDeviceUpdates() {
	topicEvents := c.node.Pubsub.Topic + eventsSubTopic
	c.Unsubscribe(topicEvents)
	close(c.updatesQueue)
	for _ = range c.updates {
		// read all remaining elements in order to close chan and go routine
	}
}

// FetchDeviceConfigs requests all device configs for the current service
func (c *ServiceClient) FetchDeviceConfigs() ([]rest.ServiceDeviceListItem, error) {
	// Get The Current Device Config
	devs, err := c.host.RequestServiceDeviceList(c.id)
	return devs, err
}

// Subscribe registers a callback for a receiving a given mqtt topic payload
func (c *ServiceClient) Subscribe(topic string, callback ServiceTopicHandler) error {
	return c.subscribe(topic, func(topic string, payload []byte) {
		callback(c, topic, payload)
	})
}

// Unsubscribe deregisters a callback for a given mqtt topic
func (c *ServiceClient) Unsubscribe(topic string) error {
	return c.unsubscribe(topic)
}

// Publish publishes a payload to a given mqtt topic
func (c *ServiceClient) Publish(topic string, payload []byte) error {
	return c.publish(topic, payload)
}

// GetProperties returns the full service properties key/value mapping
func (c *ServiceClient) GetProperties() map[string]string {
	return c.node.Properties
}

// GetProperty fetches the service property associated with key. If it does
// not exist the blank string is returned.
func (c *ServiceClient) GetProperty(key string) string {
	value, ok := c.node.Properties[key]
	if ok {
		return value
	}
	return ""
}
