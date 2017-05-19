// Package service provides the management library for a long running service
package framework

import (
	"errors"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/openchirp/framework"
)

var ErrNotImplemented = errors.New("This method is not implemented yet")

type Service struct {
	host framework.Host
	mqtt MQTT.Client
}

// CreateService creates the named service on the framework server
// and returns serviceid upon sucess
func CreateService(host framework.Host, name string) (string, error) {
	host = host // exercise that host variable
	name = name // exercise that name variable
	return "", ErrNotImplemented
}

// StartService starts the service maangement layer for service
// with id serviceid
func StartService(host framework.Host, serviceid string) (*Service, error) {
	service := new(Service)
	// we should expect mqtt settings to come from framework host
	// for now, we will simply deduce it from framework Host
	// url.Parse(host.)
	return service, nil
}

func (s *Service) GetMQTTClient() *MQTT.Client {
	return &s.mqtt
}

// Stop shuts down the
func (s *Service) Stop() {
	s.mqtt.Disconnect(0)
}

// need service go routine to listen for updates
