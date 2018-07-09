package pubsub

import (
	"fmt"
	"math/big"

	CRAND "crypto/rand"

	PahoMQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	// Sets whether AutoReconnect will be set
	defaultAutoReconnect bool = true
	disconnectWaitMS     uint = 300
)

type MQTTClient struct {
	mqtt               PahoMQTT.Client
	defaultQoS         MQTTQoS
	defaultPersistence bool
}

type MQTTQoS byte

const (
	QoSAtMostOnce  = MQTTQoS(0)
	QoSAtLeastOnce = MQTTQoS(1)
	QoSExactlyOnce = MQTTQoS(2)
	QoSUnknown     = MQTTQoS(0xFF)
)

func ParseMQTTQoS(QoS string) MQTTQoS {
	switch QoS {
	case "QoSAtMostOnce", "0":
		return QoSAtMostOnce
	case "QoSAtLeastOnce", "1":
		return QoSAtLeastOnce
	case "QoSExactlyOnce", "2":
		return QoSExactlyOnce
	default:
		return QoSUnknown
	}
}

// GenMQTTClientID generates a random client id for mqtt
func GenMQTTClientID(prefix string) (string, error) {
	r, err := CRAND.Int(CRAND.Reader, new(big.Int).SetInt64(100000))
	if err != nil {
		return "", fmt.Errorf("Failed to generate MQTT client ID: %v", err)
	}
	return prefix + r.String(), nil
}

// NewMQTTClient creates and connects an MQTT client that implements the
// PubSub interface
func NewMQTTClient(
	brokerURI, user, pass string,
	defaultQoS MQTTQoS,
	defaultPersistence bool) (*MQTTClient, error) {

	c := new(MQTTClient)
	c.defaultQoS = defaultQoS
	c.defaultPersistence = defaultPersistence

	/* Generate random client id for MQTT */
	clientID, err := GenMQTTClientID("client")
	if err != nil {
		return nil, err
	}

	/* Connect the MQTT connection */
	opts := PahoMQTT.NewClientOptions().AddBroker(brokerURI)
	opts.SetClientID(clientID)
	// http://www.hivemq.com/blog/mqtt-security-fundamentals-authentication-username-password:
	//   "The spec also states that a username without password is possible.
	//    It’s not possible to just send a password without username."
	if len(user) > 0 {
		// we do not allow absent passwords yet
		opts.SetUsername(user).SetPassword(pass)
	}
	opts.SetAutoReconnect(defaultAutoReconnect)

	/* Create and start a client using the above ClientOptions */
	c.mqtt = PahoMQTT.NewClient(opts)
	if token := c.mqtt.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return c, nil
}

// NewMQTTBridgeClient creates and connects an MQTT client that implements the
// PubSub interface. This special variant will indicate to the broker that you
// are operating as a MQTT bridge. In this case, you will not receive an echo
// of messages you publish to a topic you have subscribed to.
// Note, this is not an official MQTT feature and is only supported by a few
// brokers.
// Checkout https://github.com/mqtt/mqtt.github.io/wiki/bridge_protocol
// for more info.
func NewMQTTBridgeClient(
	brokerURI, user, pass string,
	defaultQoS MQTTQoS,
	defaultPersistence bool) (*MQTTClient, error) {

	c := new(MQTTClient)
	c.defaultQoS = defaultQoS
	c.defaultPersistence = defaultPersistence

	/* Generate random client id for MQTT */
	clientID, err := GenMQTTClientID("bridge")
	if err != nil {
		return nil, err
	}

	/* Connect the MQTT connection */
	opts := PahoMQTT.NewClientOptions().AddBroker(brokerURI)
	opts.SetClientID(clientID)
	// http://www.hivemq.com/blog/mqtt-security-fundamentals-authentication-username-password:
	//   "The spec also states that a username without password is possible.
	//    It’s not possible to just send a password without username."
	if len(user) > 0 {
		// we do not allow absent passwords yet
		opts.SetUsername(user).SetPassword(pass)
	}
	opts.SetAutoReconnect(defaultAutoReconnect)
	opts.SetProtocolVersion(4 | 0x80) // indicate bridge

	/* Create and start a client using the above ClientOptions */
	c.mqtt = PahoMQTT.NewClient(opts)
	if token := c.mqtt.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return c, nil
}

func (c *MQTTClient) Disconnect() {
	c.mqtt.Disconnect(disconnectWaitMS)
}

func (c *MQTTClient) Subscribe(topic string, callback func(topic string, payload []byte)) error {
	token := c.mqtt.Subscribe(topic, byte(c.defaultQoS), func(client PahoMQTT.Client, msg PahoMQTT.Message) {
		callback(msg.Topic(), msg.Payload())
	})
	token.Wait()
	return token.Error()
}

func (c *MQTTClient) Unsubscribe(topics ...string) error {
	token := c.mqtt.Unsubscribe(topics...)
	token.Wait()
	return token.Error()
}

func (c *MQTTClient) Publish(topic string, payload interface{}) error {
	token := c.mqtt.Publish(topic, byte(c.defaultQoS), c.defaultPersistence, payload)
	token.Wait()
	return token.Error()
}
