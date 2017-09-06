package framework

// UserClient represents the context for a single user client session
type UserClient struct {
	Client
}

// UserClientTopicHandler is a function prototype for a subscribed topic callback
type UserClientTopicHandler func(client *UserClient, topic string, payload []byte)

// StartUserClient starts the user client management layer
func StartUserClient(frameworkuri, brokeruri, id, token string) (*UserClient, error) {
	c := new(UserClient)
	err := c.startClient(frameworkuri, brokeruri, id, token)
	return c, err
}

// StopClient shuts down a started user client
func (c *UserClient) StopClient() {
	c.stopClient()
}

// Subscribe registers a callback for a receiving a given mqtt topic payload
func (c *UserClient) Subscribe(topic string, callback UserClientTopicHandler) error {
	return c.subscribe(topic, func(topic string, payload []byte) {
		callback(c, topic, payload)
	})
}

// Unsubscribe deregisters a callback for a given mqtt topics
func (c *UserClient) Unsubscribe(topics ...string) error {
	return c.unsubscribe(topics...)
}

// Publish publishes a payload to a given mqtt topic
func (c *UserClient) Publish(topic string, payload interface{}) error {
	return c.publish(topic, payload)
}
