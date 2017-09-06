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

	// Start Client
	err = c.startClient(frameworkuri, brokeruri, id, token)
	if err != nil {
		return nil, err
	}

	// Get Our Device Info
	c.node, err = c.host.RequestDeviceInfo(c.id)

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

// Unsubscribe deregisters a callback for a given mqtt topics
func (c *DeviceClient) Unsubscribe(subtopics ...string) error {
	for i, subtopic := range subtopics {
		subtopics[i] = c.node.Pubsub.Topic + "/" + subtopic
	}
	return c.unsubscribe(subtopics...)
}

// Publish publishes a payload to a given mqtt topic
func (c *DeviceClient) Publish(subtopic string, payload interface{}) error {
	return c.publish(c.node.Pubsub.Topic+"/"+subtopic, payload)
}
