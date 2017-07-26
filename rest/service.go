package rest

// ServiceNode is a container for Service Node object received
// from the RESTful JSON interface
type ServiceNode struct {
	NodeDescriptor                   // Node descriptor of Service Node
	Description    string            `json:"description"`
	Properties     map[string]string `json:"properties"`
}

/*
Services Device Config Requests Look Like The Following:
[
  {
    "id": "592c8a627d6ec25f901d9687",
    "type": "device",
    "service_config": [
      {
        "key": "DevEUI",
        "value": "test1"
      },
      {
        "key": "AppEUI",
        "value": "test2"
      },
      {
        "key": "AppKey",
        "value": "test3"
      }
    ]
  }
]
*/

type KeyValuePair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ServiceDeviceListItem represents the device and service configuration pair
// found in a Service Node's device list
type ServiceDeviceListItem struct {
	Id     string         `json:"id"`
	Config []KeyValuePair `json:"config"`
}

func (i ServiceDeviceListItem) GetID() string {
	return i.Id
}

func (i ServiceDeviceListItem) GetConfigMap() map[string]string {
	m := make(map[string]string)
	for _, v := range i.Config {
		m[v.Key] = v.Value
	}
	return m
}
