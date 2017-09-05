// Package service provides the management library for a long running service
package framework

import (
	"log"
	"math/big"

	CRAND "crypto/rand"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/openchirp/framework/rest"
)

// ClientTopicHandler is a function prototype for a subscribed topic callback
type ClientTopicHandler func(client *Client, topic string, payload []byte)

// Client represents the context for a single client
type Client struct {
	id    string
	token string
	host  rest.Host
	mqtt  MQTT.Client
}

// genClientID generates a random client id for mqtt
func (c Client) genClientID() string {
	r, err := CRAND.Int(CRAND.Reader, new(big.Int).SetInt64(100000))
	if err != nil {
		log.Fatal("Couldn't generate a random number for MQTT client ID")
	}
	return "client" + r.String()
}

// startClient starts the client connection
func (c *Client) startClient(frameworkuri, brokeruri, id, token string) error {
	/* Setup the REST interface */
	c.host = rest.NewHost(frameworkuri)

	/* Connect the MQTT connection */
	opts := MQTT.NewClientOptions().AddBroker(brokeruri)
	opts.SetClientID(c.genClientID())
	opts.SetUsername(id)
	opts.SetPassword(token)

	/* Create and start a client using the above ClientOptions */
	c.mqtt = MQTT.NewClient(opts)
	if token := c.mqtt.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

// stopService shuts down a started client
func (c *Client) stopClient() {
	c.mqtt.Disconnect(0)
}

// subscribe registers a callback for a receiving a given mqtt topic payload
func (c *Client) subscribe(topic string, callback ClientTopicHandler) error {
	token := c.mqtt.Subscribe(topic, byte(mqttQos), func(client MQTT.Client, message MQTT.Message) {
		callback(c, message.Topic(), message.Payload())
	})
	token.Wait()
	return token.Error()
}

// unsubscribe deregisters a callback for a given mqtt topic
func (c *Client) unsubscribe(topic string) error {
	token := c.mqtt.Unsubscribe(topic)
	token.Wait()
	return token.Error()
}

// publish publishes a payload to a given mqtt topic
func (c *Client) publish(topic string, payload []byte) error {
	token := c.mqtt.Publish(topic, byte(mqttQos), mqttPersistence, payload)
	token.Wait()
	return token.Error()
}

// GetMQTTClient bypasses the service interface and provies the underlying
// mqtt client context
// This will be removed in the near future
func (s *Client) GetMQTTClient() *MQTT.Client {
	return &s.mqtt
}
