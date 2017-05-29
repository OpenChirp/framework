// December 14, 2016
// Craig Hesling <craig@hesling.com>

// Package framework provides the data structures and primitive mechanisms
// for representing and communicating framework constructs with the RESTful
// server.
package rest

import (
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
	fmt.Println(uri)
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
	req, err := http.NewRequest("GET", uri, nil)
	req.SetBasicAuth(host.user, host.pass)

	// resp, err := http.Get(host.uri + servicesSubPath + "/" + serviceid)
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
	resp, err := http.Get(host.uri + deviceSubPath + "/" + deviceid)
	if err != nil {
		// should report auth problems here in future
		return deviceNode, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&deviceNode)
	return deviceNode, err
}

type PubSub struct {
	Protocol  string          `json:"protocol"`
	Topic     string          `json:"endpoint"`
	ExtraJunk json.RawMessage `json:"serviceconfig"`
}

// NodeDescriptor provides the common fields that Device and Service nodes share
type NodeDescriptor struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Pubsub PubSub `json:"pubsub"`
}
