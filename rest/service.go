package rest

// ServiceNode is a container for Service Node object received
// from the RESTful JSON interface
type ServiceNode struct {
	NodeDescriptor                   // Node descriptor of Service Node
	Description    string            `json:"description"`
	Properties     map[string]string `json:"properties"`
}

// ServiceDeviceListItem represents the device and service configuration pair
// found in a Service Node's device list
type ServiceDeviceListItem struct {
	Id     string `json:"id"`
	Config []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"service_config"`
}
