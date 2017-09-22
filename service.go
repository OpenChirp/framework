package framework

import (
	"errors"
	"fmt"
	"sync"

	"encoding/json"

	"github.com/openchirp/framework/rest"
)

const (
	eventsSubTopic         = "/thing/events"
	deviceStatusSubTopic   = "/status"
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
var ErrDeviceUpdatesAlreadyStarted = errors.New("Device updates channel already started")
var ErrDeviceUpdatesNotStarted = errors.New("Device updates channel not started")

// DeviceUpdateType represents enumeration of DeviceUpdate types
type DeviceUpdateType int

const (
	// DeviceUpdateAdd indicates that a new device linked in this service
	DeviceUpdateTypeAdd DeviceUpdateType = iota
	// DeviceUpdateRem indicates a device has unlinked this service
	DeviceUpdateTypeRem
	// DeviceUpdateUpd indicates that a device this service's config
	DeviceUpdateTypeUpd
	// DeviceUpdateTypeErr indicates an error was encountered while receiving
	// a device update event. The error message can be fetched from
	// DeviceUpdate.Error()
	DeviceUpdateTypeErr
)

// String associates a pretty name with the DeviceUpdateTypes
func (dut DeviceUpdateType) String() (s string) {
	switch dut {
	case DeviceUpdateTypeAdd:
		s = "Add"
	case DeviceUpdateTypeRem:
		s = "Remove"
	case DeviceUpdateTypeUpd:
		s = "Update"
	case DeviceUpdateTypeErr:
		s = "Error"
	}
	return
}

// DeviceUpdate represents a pending service config change for a device
type DeviceUpdate struct {
	Type   DeviceUpdateType
	Id     string
	Config map[string]string
}

func (du DeviceUpdate) Error() string {
	if du.Type == DeviceUpdateTypeErr {
		return du.Id
	}
	return ""
}

// String provides a human parsable string for DeviceUpdates
func (du DeviceUpdate) String() string {
	return fmt.Sprintf("Type: %v, Id: %s, Config: %v", du.Type, du.Id, du.Config)
}

// ServiceTopicHandler is a function prototype for a subscribed topic callback
type ServiceTopicHandler func(client *ServiceClient, topic string, payload []byte)

// ServiceClient hold a single ses.Publish(s.)rvice context
type ServiceClient struct {
	Client
	node           rest.ServiceNode
	updatesWg      sync.WaitGroup
	updatesRunning bool
	updatesQueue   chan DeviceUpdate
	updates        chan DeviceUpdate
}

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

type serviceUpdatesEncapsulation struct {
	Action string                     `json:"action"`
	Device rest.ServiceDeviceListItem `json:"thing"`
}

type serviceStatus struct {
	Message string `json:"message"`
}

type serviceDeviceStatus struct {
	Device struct {
		Id      string `json:"id"`
		Message string `json:"message"`
	} `json:"thing"`
}

// StartServiceClient starts the service management layer
func StartServiceClient(frameworkuri, brokeruri, id, token string) (*ServiceClient, error) {
	c, err := StartServiceClientStatus(frameworkuri, brokeruri, id, token, "")
	return c, err
}

// StartServiceClientStatus starts the service management layer with a optional
// statusmsg if the service disconnects improperly
func StartServiceClientStatus(frameworkuri, brokeruri, id, token, statusmsg string) (*ServiceClient, error) {
	var err error

	c := new(ServiceClient)

	// Start enough of the client manually to get REST working
	c.setAuth(id, token)
	err = c.startREST(frameworkuri)
	if err != nil {
		return nil, err
	}

	// Get Our Service Info
	c.node, err = c.host.RequestServiceInfo(c.id)
	if err != nil {
		return nil, err
	}

	// Setup will'ed status
	if statusmsg != "" {
		var msg serviceStatus
		msg.Message = statusmsg
		payload, err := json.Marshal(&msg)
		if err != nil {
			return nil, ErrMarshalStatusMessage
		}
		c.setWill(c.node.Pubsub.Topic+statusSubTopic, []byte(payload))
	}

	// Start MQTT
	err = c.startMQTT(brokeruri)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// StopClient shuts down a started service
func (c *ServiceClient) StopClient() {
	c.stopClient()
}

// SetStatus publishes the service status message
func (c *ServiceClient) SetStatus(msgs ...interface{}) error {
	var statusmsg serviceStatus
	statusmsg.Message = fmt.Sprint(msgs...)
	payload, err := json.Marshal(&statusmsg)
	if err != nil {
		return ErrMarshalStatusMessage
	}
	return c.Publish(c.node.Pubsub.Topic+statusSubTopic, payload)
}

// SetDeviceStatus publishes a device's linked service status message
func (c *ServiceClient) SetDeviceStatus(id string, msgs ...interface{}) error {
	var statusmsg serviceDeviceStatus
	statusmsg.Device.Id = id
	statusmsg.Device.Message = fmt.Sprint(msgs...)
	payload, err := json.Marshal(&statusmsg)
	if err != nil {
		return ErrMarshalDeviceStatusMessage
	}
	return c.Publish(c.node.Pubsub.Topic+deviceStatusSubTopic, payload)
}

func (c *ServiceClient) updateEventsHandler() func(topic string, payload []byte) {
	return func(topic string, payload []byte) {
		c.updatesWg.Add(1)
		defer c.updatesWg.Done()
		if c.updatesRunning {
			// action: new, update, delete
			var mqttMsg serviceUpdatesEncapsulation
			var devUpdate DeviceUpdate

			err := json.Unmarshal(payload, &mqttMsg)
			if err != nil {
				c.updatesQueue <- DeviceUpdate{
					Type: DeviceUpdateTypeErr,
					Id:   fmt.Sprintf("Failed to unmarshal message on topic %s\n", topic),
				}
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
		}
	}
}

func (c *ServiceClient) startDeviceUpdatesQueue() error {
	/* Setup MQTT based device updates to feed updatesQueue */
	topicEvents := c.node.Pubsub.Topic + eventsSubTopic
	if c.updatesRunning {
		return ErrDeviceUpdatesAlreadyStarted
	}
	c.updatesRunning = true
	c.updatesQueue = make(chan DeviceUpdate, deviceUpdatesBuffering)
	err := c.Subscribe(topicEvents, c.updateEventsHandler())
	if err != nil {
		c.stopDeviceUpdatesQueue()
		return err
	}
	return nil
}

func (c *ServiceClient) stopDeviceUpdatesQueue() error {
	topicEvents := c.node.Pubsub.Topic + eventsSubTopic
	if c.updatesRunning {
		return ErrDeviceUpdatesNotStarted
	}

	c.Unsubscribe(topicEvents)
	c.updatesRunning = false

	// Unblock all possible updateEventsHandlers while we wait
	go func() {
		for _ = range c.updatesQueue {
			// read all remaining elements in order to close chan and go routines
		}
		c.updatesQueue = nil
	}()
	// wait for all activivley running routines to finish writing to channel
	c.updatesWg.Wait()
	close(c.updatesQueue)
	return nil
}

// StartDeviceUpdatesSimple subscribes to the live mqtt service news topic and opens
// a channel to read the updates from. It will automatically fetch the initial
// configuration and send those as DeviceUpdateTypeAdd updates first.
// Due to the time between subscribing to live events and requesting the static
// configuration, there may be redundant DeviceUpdateTypeAdd updates. Your
// program should account for this.
func (c *ServiceClient) StartDeviceUpdatesSimple() (<-chan DeviceUpdate, error) {

	/* Setup MQTT based device updates to feed updatesQueue */
	err := c.startDeviceUpdatesQueue()
	if err != nil {
		return nil, err
	}

	/* Preload device updates from REST request */
	configUpdates, err := c.FetchDeviceConfigsAsUpdates()
	if err != nil {
		c.stopDeviceUpdatesQueue()
		return nil, err
	}
	c.updates = make(chan DeviceUpdate, len(configUpdates))
	for _, update := range configUpdates {
		c.updates <- update
	}

	/* Connect updatesQueue channel to updates channel */
	go func() {
		for update := range c.updatesQueue {
			c.updates <- update
		}
		close(c.updates)
	}()

	return c.updates, err
}

// StartDeviceUpdates subscribes to the live service events topic and opens
// a channel to read the updates from. This does not inject the initial
// configurations into the channel at start like StartDeviceUpdatesSimple.
func (c *ServiceClient) StartDeviceUpdates() (<-chan DeviceUpdate, error) {

	/* Setup MQTT based device updates to feed updatesQueue */
	err := c.startDeviceUpdatesQueue()
	if err != nil {
		return nil, err
	}

	/* Make the updates channel */
	c.updates = make(chan DeviceUpdate)

	/* Connect updatesQueue channel to updates channel */
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

// FetchDeviceConfigsAsUpdates requests all device configs for the current
// service and converts them into DeviceUpdate with DeviceUpdateTypeAdd as the
// type
func (c *ServiceClient) FetchDeviceConfigsAsUpdates() ([]DeviceUpdate, error) {
	// Get The Current Device Config
	deviceConfigs, err := c.host.RequestServiceDeviceList(c.id)
	if err != nil {
		return nil, err
	}
	updates := make([]DeviceUpdate, len(deviceConfigs))
	for i, devConfig := range deviceConfigs {
		updates[i] = DeviceUpdate{
			Type:   DeviceUpdateTypeAdd,
			Id:     devConfig.Id,
			Config: devConfig.GetConfigMap(),
		}
	}
	return updates, nil
}

// Subscribe registers a callback for a receiving a given mqtt topic payload
func (c *ServiceClient) Subscribe(topic string, callback func(topic string, payload []byte)) error {
	return c.subscribe(topic, callback)
}

// SubscribeWithClient registers a callback for a receiving a given mqtt
// topic payload and provides the client object
func (c *ServiceClient) SubscribeWithClient(topic string, callback ServiceTopicHandler) error {
	return c.subscribe(topic, func(topic string, payload []byte) {
		callback(c, topic, payload)
	})
}

// Unsubscribe deregisters a callback for a given mqtt topic
func (c *ServiceClient) Unsubscribe(topics ...string) error {
	return c.unsubscribe(topics...)
}

// Publish publishes a payload to a given mqtt topic
func (c *ServiceClient) Publish(topic string, payload interface{}) error {
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
