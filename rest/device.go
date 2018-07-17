// December 14, 2016
// Craig Hesling <craig@hesling.com>

package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// DeviceListServiceItem represents the service and service configuration pair
// found in in a Device Node's service list
type DeviceListServiceItem struct {
	ServiceID     string         `json:"service_id"`
	ServiceConfig []KeyValuePair `json:"config,omitempty"`
}

// DeviceNode is a container for Device Node object received
// from the RESTful JSON interface
type DeviceNode struct {
	NodeDescriptor                         // Node descriptor of Device Node
	Properties     map[string]string       `json:"properties"`
	Services       []DeviceListServiceItem `json:"linked_services"`
}

// RequestDeviceInfo makes an HTTP GET to the framework server requesting
// the Device Node information for the device with ID deviceID.
func (host Host) RequestDeviceInfo(deviceID string) (DeviceNode, error) {
	var deviceNode DeviceNode
	uri := host.uri + rootAPISubPath + deviceSubPath + "/" + deviceID
	fmt.Println("DevURI:", uri)
	req, err := http.NewRequest("GET", uri, nil)
	req.SetBasicAuth(host.user, host.pass)

	resp, err := host.client.Do(req)
	if err != nil {
		// should report auth problems here in future
		return deviceNode, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&deviceNode)
	return deviceNode, err
}

// RequestLinkedService makes an HTTP POST to the framework server to link the
// specified serviceID to device deviceID.
func (host Host) RequestLinkedService(deviceID, serviceID string) (DeviceListServiceItem, error) {
	var deviceServiceItem DeviceListServiceItem
	uri := host.uri + rootAPISubPath + deviceSubPath + "/" + deviceID + "/service/" + serviceID
	req, err := http.NewRequest("GET", uri, nil)
	req.SetBasicAuth(host.user, host.pass)

	resp, err := host.client.Do(req)
	if err != nil {
		return deviceServiceItem, err
	}
	defer resp.Body.Close()
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
	req.SetBasicAuth(host.user, host.pass)

	// resp, err := http.Get(uri)
	resp, err := host.client.Do(req)
	if err != nil {
		resp.Body.Close()
	}
	return err
}

// DelinkService makes an HTTP DELETE to the framework server to delink the
// specified serviceID from device deviceID.
func (host Host) DelinkService(deviceID, serviceID string) error {
	uri := host.uri + rootAPISubPath + deviceSubPath + "/" + deviceID + "/service/" + serviceID
	req, err := http.NewRequest("DELETE", uri, nil)
	req.SetBasicAuth(host.user, host.pass)

	// resp, err := http.Get(uri)
	resp, err := host.client.Do(req)
	if err != nil {
		resp.Body.Close()
	}
	return err
}

// ExecuteCommand makes an HTTP POST to the framework server to execute the
// specified commandID on device deviceID.
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
