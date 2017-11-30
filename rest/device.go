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
	ServiceConfig []KeyValuePair `json:"config"`
}

// DeviceNode is a container for Device Node object received
// from the RESTful JSON interface
type DeviceNode struct {
	NodeDescriptor                         // Node descriptor of Device Node
	Properties     map[string]string       `json:"properties"`
	Services       []DeviceListServiceItem `json:"linked_services"`
}

// RequestDeviceInfo makes an HTTP GET to the framework server requesting
// the Device Node information for the device with ID deviceid.
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

// ExecuteCommand makes an HTTP POST to the framework server to execute the
// specified commmandID on device deviceID.
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
