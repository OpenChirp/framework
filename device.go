package framework

import (
	"github.com/openchirp/framework/rest"
)

// DeviceClient represents the context for a single user device session
type DeviceClient struct {
	Client
	node rest.DeviceNode
}

// StartDeviceClient starts the device client management layer
func StartDeviceClient(frameworkuri, brokeruri, id, token string) (*DeviceClient, error) {
	var err error
	c := new(DeviceClient)

	// Get Our Device Info
	c.node, err = c.host.RequestDeviceInfo(c.id)
	if err != nil {
		return nil, err
	}

	// Start Client
	err = c.startClient(frameworkuri, brokeruri, id, token)
	return c, err
}

// StopClient shuts down a started device client
func (c *DeviceClient) StopClient() {
	c.stopClient()
}

// Subscribe registers a callback for receiving on a device subtopic
func (c *DeviceClient) Subscribe(subtopic string, callback ClientTopicHandler) error {
	return c.subscribe(c.node.Pubsub.Topic+"/"+subtopic, callback)
}

// Unsubscribe deregisters a callback for a given mqtt topic
func (c *DeviceClient) Unsubscribe(subtopic string) error {
	return c.unsubscribe(c.node.Pubsub.Topic + "/" + subtopic)
}

// Publish publishes a payload to a given mqtt topic
func (c *DeviceClient) Publish(subtopic string, payload []byte) error {
	return c.publish(c.node.Pubsub.Topic+"/"+subtopic, payload)
}
