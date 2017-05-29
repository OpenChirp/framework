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

/*
openchirp/services/592880c57d6ec25f901d9668/thing/new
{
	"thing":{
		"type":"device",
		"id":"592c8a627d6ec25f901d9687",
		"config":[{"key":"DevEUI","value":"test1"},
					{"key":"AppEUI","value":"test2"},
					{"key":"AppKey","value":"test3"}]
		}
}
*/

type ServiceNewsCapsulation struct {
	Thing ServiceDeviceUpdate `json:"thing"`
}

type ServiceDeviceUpdate struct {
	Id     string `json:"id"`
	Config []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"config"`
}
