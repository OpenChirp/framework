// December 14, 2016
// Craig Hesling <craig@hesling.com>

// Package framework provides the data structures and primitive mechanisms
// for representing and communicating framework constructs with the RESTful
// server.
package framework

import (
	"encoding/json"
	"net/http"
)

const (
	deviceSubPath   = "/device"
	servicesSubPath = "/service"
)

// Host represents the RESTful HTTP server that hosts the framework
type Host struct {
	uri string
	// This is where we add APIKeys and username/password for user
}

// NewHost returns an object referencing the framework server
func NewHost(uri string) Host {
	// no need to decompose uri using net/url package
	return Host{uri}
}

// RequestServiceInfo makes an HTTP GET to the framework server requesting
// the Service Node information for service with ID serviceid.
func (host Host) RequestServiceInfo(serviceid string) (ServiceNode, error) {
	var serviceNode ServiceNode
	resp, err := http.Get(host.uri + servicesSubPath + "/" + serviceid)
	if err != nil {
		// should report auth problems here in future
		return serviceNode, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&serviceNode)
	return serviceNode, err
}

// RequestDeviceInfo makes an HTTP GET to the framework server requesting
// the Device Node information for device with ID deviceid.
func (host Host) RequestDeviceInfo(deviceid string) (DeviceNode, error) {
	var deviceNode DeviceNode
	resp, err := http.Get(host.uri + deviceSubPath + "/" + deviceid)
	if err != nil {
		// should report auth problems here in future
		return deviceNode, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&deviceNode)
	return deviceNode, err
}

// NodeDescriptor provides the common fields that Device and Service nodes share
type NodeDescriptor struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Pubsub PubSub `json:"pubsub"`
}

// ServiceListDeviceItem represents the device and service configuration pair
// found in a Service Node's device list
type ServiceListDeviceItem struct {
	NodeDescriptor                 // Node descriptor listed Device Node
	ServiceConfig  json.RawMessage `json:"serviceconfig"`
}

// ServiceNode is a container for Service Node object received
// from the RESTful JSON interface
type ServiceNode struct {
	NodeDescriptor                         // Node descriptor of Service Node
	Description    string                  `json:"description"`
	DeviceNodes    []ServiceListDeviceItem `json:"devicenodes"`
}

// DeviceListServiceItem represents the service and service configuration pair
// found in in a Device Node's service list
type DeviceListServiceItem struct {
	ServiceID     string          `json:"serviceid"`
	ServiceConfig json.RawMessage `json:"serviceconfig"`
}

// DeviceNode is a container for Device Node object received
// from the RESTful JSON interface
type DeviceNode struct {
	NodeDescriptor                         // Node descriptor of Device Node
	Data           map[string]interface{}  `json:"data"`
	Properties     map[string]interface{}  `json:"properties"`
	Services       []DeviceListServiceItem `json:"services"`
}
