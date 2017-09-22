// December 14, 2016
// Craig Hesling <craig@hesling.com>

// Package rest provides the data structures and primitive mechanisms
// for representing and communicating framework constructs with the RESTful
// server.
package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	rootAPISubPath        = "/api"
	deviceSubPath         = "/device"
	servicesSubPath       = "/service"
	serviceDevicesSubPath = "/things"
)

// Host represents the RESTful HTTP server that hosts the framework
type Host struct {
	uri string
	// This is where we add APIKeys and username/password for user
	user   string
	pass   string
	client http.Client
}

// NewHost returns an object referencing the framework server
func NewHost(uri string) Host {
	// no need to decompose uri using net/url package
	return Host{uri: uri, client: http.Client{}}
}

func (host *Host) Login(username, password string) error {
	host.user = username
	host.pass = password
	// TODO: Check login credentials -- return error if no good
	return nil
}

// RequestServiceInfo makes an HTTP GET to the framework server requesting
// the Service Node information for service with ID serviceid.
func (host Host) RequestServiceInfo(serviceid string) (ServiceNode, error) {
	var serviceNode ServiceNode
	uri := host.uri + rootAPISubPath + servicesSubPath + "/" + serviceid
	req, err := http.NewRequest("GET", uri, nil)
	req.SetBasicAuth(host.user, host.pass)

	// resp, err := http.Get(host.uri + servicesSubPath + "/" + serviceid)
	resp, err := host.client.Do(req)
	if err != nil {
		// should report auth problems here in future
		return serviceNode, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&serviceNode)
	return serviceNode, err
}

// RequestServiceDeviceList
func (host Host) RequestServiceDeviceList(serviceid string) ([]ServiceDeviceListItem, error) {
	var serviceDeviceListItems = make([]ServiceDeviceListItem, 0)
	uri := host.uri + rootAPISubPath + servicesSubPath + "/" + serviceid + serviceDevicesSubPath
	fmt.Println("Service URI: ", uri)
	req, err := http.NewRequest("GET", uri, nil)
	req.SetBasicAuth(host.user, host.pass)

	resp, err := host.client.Do(req)
	if err != nil {
		// should report auth problems here in future
		return serviceDeviceListItems, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&serviceDeviceListItems)
	return serviceDeviceListItems, err
}

// RequestDeviceInfo makes an HTTP GET to the framework server requesting
// the Device Node information for device with ID deviceid.
func (host Host) RequestDeviceInfo(deviceid string) (DeviceNode, error) {
	var deviceNode DeviceNode
	uri := host.uri + rootAPISubPath + deviceSubPath + "/" + deviceid
	fmt.Println("DevURI:", uri)
	req, err := http.NewRequest("GET", uri, nil)
	req.SetBasicAuth(host.user, host.pass)

	// resp, err := http.Get(uri)
	resp, err := host.client.Do(req)
	if err != nil {
		// should report auth problems here in future
		return deviceNode, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&deviceNode)
	return deviceNode, err
}

// RequestDeviceInfo makes an HTTP GET to the framework server requesting
// the Device Node information for device with ID deviceid.
func (host Host) ExecuteCommand(deviceID, commandID string) error {
	uri := host.uri + rootAPISubPath + deviceSubPath + "/" + deviceID + "/command/" + commandID
	req, err := http.NewRequest("POST", uri, bytes.NewReader([]byte("{}")))
	req.SetBasicAuth(host.user, host.pass)

	// resp, err := http.Get(uri)
	resp, err := host.client.Do(req)
	if err != nil {
		resp.Body.Close()
	}
	return err
}

// PubSub describes a node's pubsub endpoint
type PubSub struct {
	Protocol  string          `json:"protocol"`
	Topic     string          `json:"endpoint"`
	ExtraJunk json.RawMessage `json:"serviceconfig"`
}

// Owner describes the owning user's details
type Owner struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// NodeDescriptor provides the common fields that Device and Service nodes share
type NodeDescriptor struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Pubsub PubSub `json:"pubsub"`
	Owner  Owner  `json:"owner"`
}
