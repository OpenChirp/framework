// December 14, 2016
// Craig Hesling <craig@hesling.com>

package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// DeviceListServiceItem represents the service and service configuration pair
// found in in a Device Node's service list
type DeviceListServiceItem struct {
	ServiceID     string         `json:"service_id"`
	ServiceConfig []KeyValuePair `json:"config,omitempty"`
}

// TransducerInfo describes a transducer within a Device.
type TransducerInfo struct {
	Name       string `json:"name"`
	Unit       string `json:"unit"`
	IsActuable bool   `json:"is_actuable"`
}

// TransducerValue holds a transducer description with a single value
// and timestamp.
type TransducerValue struct {
	TransducerInfo
	Value          string    `json:"value"`
	ValueTimestamp time.Time `json:"timestamp"`
}

// DeviceNode is a container for Device Node object received
// from the RESTful JSON interface
type DeviceNode struct {
	NodeDescriptor                         // Node descriptor of Device Node
	Properties     map[string]string       `json:"properties"`
	Transducers    []TransducerInfo        `json:"transducers"`
	Services       []DeviceListServiceItem `json:"linked_services"`
}

// Clone creates a new copy of the DeviceNode.
// This is necessary because a DeviceNode has embedded slices.
func (n *DeviceNode) Clone() DeviceNode {
	var ret = *n
	ret.Transducers = []TransducerInfo{}
	ret.Services = []DeviceListServiceItem{}
	copy(ret.Transducers, n.Transducers)
	copy(ret.Services, n.Services)
	return ret
}

// DeviceAll makes an HTTP GET to the framework server requesting
// the a list of all devices
func (host Host) DeviceAll() ([]NodeDescriptor, error) {
	var devices []NodeDescriptor
	uri := host.uri + rootAPISubPath + deviceSubPath
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return devices, err
	}
	req.SetBasicAuth(host.user, host.pass)

	resp, err := host.client.Do(req)
	if err != nil {
		// should report auth problems here in future
		return devices, err
	}
	defer resp.Body.Close()
	if err := DecodeOCError(resp); err != nil {
		return devices, err
	}
	err = json.NewDecoder(resp.Body).Decode(&devices)
	return devices, err
}

// RequestDeviceInfo makes an HTTP GET to the framework server requesting
// the Device Node information for the device with ID deviceID.
func (host Host) RequestDeviceInfo(deviceID string) (DeviceNode, error) {
	var deviceNode DeviceNode
	uri := host.uri + rootAPISubPath + deviceSubPath + "/" + deviceID
	fmt.Println("DevURI:", uri)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return deviceNode, err
	}
	req.SetBasicAuth(host.user, host.pass)

	resp, err := host.client.Do(req)
	if err != nil {
		// should report auth problems here in future
		return deviceNode, err
	}
	defer resp.Body.Close()
	if err := DecodeOCError(resp); err != nil {
		return deviceNode, err
	}
	err = json.NewDecoder(resp.Body).Decode(&deviceNode)
	return deviceNode, err
}

// DeviceTransducerValues makes an HTTP GET to the framework server requesting
// the transducers last value list for the device with ID deviceID.
func (host Host) DeviceTransducerValues(deviceID string) ([]TransducerValue, error) {
	var transducers []TransducerValue
	uri := host.uri + rootAPISubPath + deviceSubPath + "/" + deviceID + "/transducer"
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return transducers, err
	}
	req.SetBasicAuth(host.user, host.pass)

	resp, err := host.client.Do(req)
	if err != nil {
		// should report auth problems here in future
		return transducers, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&transducers)
	return transducers, err
}

// DeviceTransducerLastValue makes an HTTP GET to the framework server requesting
// the transducers last value for the device with ID deviceID and transducer
// with with ID or name transducerID.
func (host Host) DeviceTransducerLastValue(deviceID, transducerID string) ([]byte, error) {
	var value []byte
	uri := host.uri + rootAPISubPath + deviceSubPath + "/" + deviceID + "/transducer/" + transducerID
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return value, err
	}
	req.SetBasicAuth(host.user, host.pass)

	resp, err := host.client.Do(req)
	if err != nil {
		// should report auth problems here in future
		return value, err
	}
	defer resp.Body.Close()
	if err := DecodeOCError(resp); err != nil {
		return value, err
	}
	value, err = ioutil.ReadAll(resp.Body)
	return value, err
}

// RequestLinkedService makes an HTTP POST to the framework server to link the
// specified serviceID to device deviceID.
func (host Host) RequestLinkedService(deviceID, serviceID string) (DeviceListServiceItem, error) {
	var deviceServiceItem DeviceListServiceItem
	uri := host.uri + rootAPISubPath + deviceSubPath + "/" + deviceID + "/service/" + serviceID
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return deviceServiceItem, err
	}
	req.SetBasicAuth(host.user, host.pass)

	resp, err := host.client.Do(req)
	if err != nil {
		return deviceServiceItem, err
	}
	defer resp.Body.Close()
	if err := DecodeOCError(resp); err != nil {
		return deviceServiceItem, err
	}
	err = json.NewDecoder(resp.Body).Decode(&deviceServiceItem)
	return deviceServiceItem, err
}

// LinkService makes an HTTP POST to the framework server to link the
// specified serviceID to device deviceID.
func (host Host) LinkService(deviceID, serviceID string, config []KeyValuePair) error {
	uri := host.uri + rootAPISubPath + deviceSubPath + "/" + deviceID + "/service/" + serviceID
	body, err := json.Marshal(DeviceListServiceItem{
		ServiceID:     serviceID,
		ServiceConfig: config,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", uri, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.SetBasicAuth(host.user, host.pass)

	resp, err := host.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return DecodeOCError(resp)
}

// DelinkService makes an HTTP DELETE to the framework server to delink the
// specified serviceID from device deviceID.
func (host Host) DelinkService(deviceID, serviceID string) error {
	uri := host.uri + rootAPISubPath + deviceSubPath + "/" + deviceID + "/service/" + serviceID
	req, err := http.NewRequest("DELETE", uri, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(host.user, host.pass)

	resp, err := host.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return DecodeOCError(resp)
}

// ExecuteCommand makes an HTTP POST to the framework server to execute the
// specified commandID on device deviceID.
func (host Host) ExecuteCommand(deviceID, commandID string) error {
	uri := host.uri + rootAPISubPath + deviceSubPath + "/" + deviceID + "/command/" + commandID
	req, err := http.NewRequest("POST", uri, bytes.NewReader([]byte("{}")))
	if err != nil {
		return err
	}
	req.SetBasicAuth(host.user, host.pass)

	resp, err := host.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return DecodeOCError(resp)
}
