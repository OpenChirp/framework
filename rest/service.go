package rest

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