// December 14, 2016
// Craig Hesling <craig@hesling.com>

package rest

import (
	"encoding/json"
)

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
