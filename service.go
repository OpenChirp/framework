// Package service provides the management library for a long running service
package framework

import (
	"errors"
	"log"
	"math/big"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/openchirp/framework"
	"github.com/openchirp/framework/rest"
)

/* Options to be filled in by arguments */
var mqttBroker string
var mqttUser string
var mqttPass string
var mqttQos uint

var ErrNotImplemented = errors.New("This method is not implemented yet")

type Service struct {
	host    framework.Host
	mqtt    MQTT.Client
	node    rest.ServiceNode
	devices []rest.DeviceListServiceItem
}

/* Generate a random client id for mqtt */
func (s Service) genclientid() string {
	r, err := CRAND.Int(CRAND.Reader, new(big.Int).SetInt64(100000))
	if err != nil {
		log.Fatal("Couldn't generate a random number for MQTT client ID")
	}
	return s.node.ID + "-" + r.String()
}

// CreateService creates the named service on the framework server
// and returns serviceid upon sucess
// func CreateService(host framework.Host, name string) (string, error) {
// 	host = host // exercise that host variable
// 	name = name // exercise that name variable
// 	return "", ErrNotImplemented
// }

// StartService starts the service maangement layer for service
// with id serviceid
func StartService(host framework.Host, serviceid string) (*Service, error) {
	s := new(Service)
	// we should expect mqtt settings to come from framework host
	// for now, we will simply deduce it from framework Host
	// url.Parse(host.)

	// Get Our Service Info
	s.node = host.RequestServiceInfo(serviceid)

	// Connect to MQTT
	/* Setup basic MQTT connection */
	opts := MQTT.NewClientOptions().AddBroker(s.node.Properties["MQTTBroker"])
	opts.SetClientID(s.genclientid())
	opts.SetUsername(s.node.Properties["MQTTUser"])
	opts.SetPassword(s.node.Properties["MQTTPass"])

	/* Create and start a client using the above ClientOptions */
	s.mqtt = MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		return nil, err
	}

	// Subscribe to our news topic
	// s.mqtt.Subscribe()

	// Get The Current Device Config
	s.devices = host.RequestServiceDeviceList(serviceid)

	return s, nil
}

func (s *Service) GetMQTTClient() *MQTT.Client {
	return &s.mqtt
}

// Stop shuts down the
func (s *Service) Stop() {
	s.mqtt.Disconnect(0)
}

// need service go routine to listen for updates

func (s Service) GetProperties() map[string]string {
	return s.node.Properties
}

func (s *Service) GetDevices() {

}
