// Package service provides the management library for a long running service
package framework

import (
	"log"
	"math/big"

	CRAND "crypto/rand"

	"os"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/openchirp/framework/rest"
)

// TopicHandler is a function prototype for a subscribed topic callback
type ClientTopicHandler func(client *Client, topic string, payload []byte)

// Client represents the context for a single client
type Client struct {
	host rest.Host
	mqtt MQTT.Client
	log  *log.Logger
}

// genClientID generates a random client id for mqtt
func (c Client) genClientID() string {
	r, err := CRAND.Int(CRAND.Reader, new(big.Int).SetInt64(100000))
	if err != nil {
		log.Fatal("Couldn't generate a random number for MQTT client ID")
	}
	return "client" + r.String()
}

// CreateService creates the named service on the framework server
// and returns serviceid upon sucess
// func CreateService(host framework.Host, name string) (string, error) {
// 	host = host // exercise that host variable
// 	name = name // exercise that name variable
// 	return "", ErrNotImplemented
// }

// StartClient starts the service management layer for service
// with id serviceID
func StartClient(host rest.Host, broker, user, pass string) (*Client, error) {
	c := new(Client)
	c.host = host
	c.log = log.New(os.Stderr, "Service:", log.Flags())

	// we should expect mqtt settings to come from framework host
	// for now, we will simply deduce it from framework Host
	// url.Parse(host.)

	// Connect to MQTT
	/* Setup basic MQTT connection */
	opts := MQTT.NewClientOptions().AddBroker(broker)
	opts.SetClientID(c.genClientID())
	opts.SetUsername(user)
	opts.SetPassword(pass)

	/* Create and start a client using the above ClientOptions */
	c.mqtt = MQTT.NewClient(opts)
	if token := c.mqtt.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return c, nil
}

// StopService shuts down a started service
func (c *Client) StopClient() {
	c.mqtt.Disconnect(0)
}

// Subscribe registers a callback for a receiving a given mqtt topic payload
func (c *Client) Subscribe(topic string, callback ClientTopicHandler) error {
	token := c.mqtt.Subscribe(topic, byte(mqttQos), func(client MQTT.Client, message MQTT.Message) {
		callback(c, message.Topic(), message.Payload())
	})
	token.Wait()
	return token.Error()
}

// Unsubscribe deregisters a callback for a given mqtt topic
func (c *Client) Unsubscribe(topic string) error {
	token := c.mqtt.Unsubscribe(topic)
	token.Wait()
	return token.Error()
}

// Publish publishes a payload to a given mqtt topic
func (c *Client) Publish(topic string, payload []byte) error {
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
