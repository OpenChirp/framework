package pubsub

import (
	"math/big"

	CRAND "crypto/rand"

	PahoMQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	// Sets whether AutoReonnect will be set
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
	case "QoSAtMostOnce":
		fallthrough
	case "0":
		return QoSAtMostOnce

	case "QoSAtLeastOnce":
		fallthrough
	case "1":
		return QoSAtLeastOnce

	case "QoSExactlyOnce":
		fallthrough
	case "2":
		return QoSExactlyOnce
	default:
		return QoSUnknown
	}
}

// genClientID generates a random client id for mqtt
func (c MQTTClient) genClientID() (string, error) {
	r, err := CRAND.Int(CRAND.Reader, new(big.Int).SetInt64(100000))
	if err != nil {
		return "", err
	}
	return "client" + r.String(), nil
}

func NewMQTTClient(
	brokeruri, user, pass string,
	defaultQoS MQTTQoS,
	defaultPersistence bool) (*MQTTClient, error) {

	c := new(MQTTClient)
	c.defaultQoS = defaultQoS
	c.defaultPersistence = defaultPersistence

	/* Generate random client id for MQTT */
	clinetid, err := c.genClientID()
	if err != nil {
		return nil, err
	}

	/* Connect the MQTT connection */
	opts := PahoMQTT.NewClientOptions().AddBroker(brokeruri)
	opts.SetClientID(clinetid)
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
