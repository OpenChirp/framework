// Package framework provides the management interfaces for
// Users, Devices, and Services. Please use the appropriate top level
// class for the type of interface you need. The parent class is Client.
// 		Users - StartUserClient()
// 		Device - StartDeviceClient()
// 		Service - StartServiceClientManaged()
package framework

import (
	"log"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/openchirp/framework/pubsub"
	"github.com/openchirp/framework/rest"
)

// MQTTBridgeClient sets whether the MQTT client will identify itself as a
// bridge to the broker
var MQTTBridgeClient = false

const (
	mqttAutoReconnect        = true
	mqttQoS             byte = 2
	mqttRetained             = false
	mqttProtocolVersion uint = 4
)

// ClientTopicHandler is a function prototype for a subscribed topic callback
type ClientTopicHandler func(topic string, payload []byte)

// Client represents the context for a single client
type Client struct {
	id          string
	token       string
	host        rest.Host
	willTopic   string
	willPayload []byte
	mqtt        MQTT.Client
}

// setAuth sets basic client authentication parameters
func (c *Client) setAuth(id, token string) {
	c.id = id
	c.token = token
}

func (c *Client) startREST(frameworkURI string) error {
	c.host = rest.NewHost(frameworkURI)
	if err := c.host.Login(c.id, c.token); err != nil {
		return err
	}
	return nil
}

func (c *Client) setWill(topic string, payload []byte) {
	c.willTopic = topic
	c.willPayload = payload
}

/*
	From the documentation on NewClientOptions, ClientOptions are
	created with the following defaults:
		Port: 1883
		CleanSession: True
		Order: True
		KeepAlive: 30 (seconds)
		ConnectTimeout: 30 (seconds)
		MaxReconnectInterval 10 (minutes)
		AutoReconnect: True

	From the documentation of other ClientOptions receivers:
		SetStore will set the implementation of the Store interface used to
		         provide message persistence in cases where QoS levels
		         QoS_ONE or QoS_TWO are used. If no store is provided, then
		         the client will use MemoryStore by default.
		SetMessageChannelDepth sets the size of the internal queue that
		                       holds messages while the client is temporairily
		                       offline, allowing the application to publish
		                       when the client is reconnecting. This setting
		                       is only valid if AutoReconnect is set to true,
		                       it is otherwise ignored.
		SetPingTimeout will set the amount of time (in seconds) that
		               the client will wait after sending a PING request to
		               the broker, before deciding that the connection has
		               been lost.
		               Default is 10 seconds.
		SetWriteTimeout puts a limit on how long a mqtt
		                publish should block until it unblocks with a timeout
		                error. A duration of 0 never times out.
		                Default 30 seconds
		SetMaxReconnectInterval sets the maximum time that will be waited
		                        between reconnection attempts when connection
		                        is lost
		SetKeepAlive will set the amount of time (in seconds) that the
		             client should wait before sending a PING request
		             to the broker. This will allow the client to know
		             that a connection has not been lost with the server.
		SetConnectTimeout limits how long the client will wait when trying
		                  to open a connection to an MQTT server before
		                  imeing out and erroring the attempt. A duration
		                  of 0 never times out. Default 30 seconds.
		                  Currently only operational on TCP/TLS connections.
		SetAutoReconnect sets whether the automatic reconnection logic should
		                 be used when the connection is lost, even if disabled
		                 the ConnectionLostHandler is still called
*/
func (c *Client) startMQTT(brokerURI string) error {
	/* Connect the MQTT connection */
	opts := MQTT.NewClientOptions().AddBroker(brokerURI)

	var prefix = "client"
	opts.SetProtocolVersion(mqttProtocolVersion)
	if MQTTBridgeClient {
		prefix = "bridge"
		opts.SetProtocolVersion(mqttProtocolVersion | 0x80)
	}
	clientID, err := pubsub.GenMQTTClientID(prefix)
	if err != nil {
		log.Fatal(err)
	}
	opts.SetClientID(clientID)
	opts.SetUsername(c.id).SetPassword(c.token)
	opts.SetAutoReconnect(mqttAutoReconnect)
	if c.willTopic != "" {
		opts.SetBinaryWill(c.willTopic, c.willPayload, mqttQoS, mqttRetained)
	}

	/* Create and start a client using the above ClientOptions */
	c.mqtt = MQTT.NewClient(opts)
	if token := c.mqtt.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// startClient sets auth, starts REST, and starts MQTT
func (c *Client) startClient(frameworkURI, brokerURI, id, token string) error {
	/* Setup basic client parameters */
	c.setAuth(id, token)

	/* Setup the REST interface */
	err := c.startREST(frameworkURI)
	if err != nil {
		return err
	}

	return c.startMQTT(brokerURI)
}

// stopService shuts down a started client
func (c *Client) stopClient() {
	c.mqtt.Disconnect(0)
}

// subscribe registers a callback for a receiving a given mqtt topic payload
func (c *Client) subscribe(topic string, callback ClientTopicHandler) error {
	token := c.mqtt.Subscribe(topic, byte(mqttQos), func(client MQTT.Client, message MQTT.Message) {
		callback(message.Topic(), message.Payload())
	})
	token.Wait()
	return token.Error()
}

// unsubscribe deregisters a callback for a given mqtt topics
func (c *Client) unsubscribe(topics ...string) error {
	token := c.mqtt.Unsubscribe(topics...)
	token.Wait()
	return token.Error()
}

// publish publishes a payload to a given mqtt topic
func (c *Client) publish(topic string, payload interface{}) error {
	token := c.mqtt.Publish(topic, byte(mqttQos), mqttPersistence, payload)
	token.Wait()
	return token.Error()
}

// FetchDeviceInfo requests and fetches device information from the REST interface
func (s *Client) FetchDeviceInfo(deviceID string) (rest.DeviceNode, error) {
	d, err := s.host.RequestDeviceInfo(deviceID)
	return d, err
}
