// December 14, 2016
// Craig Hesling <craig@hesling.com>

package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ServiceNode is a container for Service Node object received
// from the RESTful JSON interface
type ServiceNode struct {
	NodeDescriptor                            // Node descriptor of Service Node
	Description      string                   `json:"description"`
	Properties       map[string]string        `json:"properties"`
	ConfigParameters []ServiceConfigParameter `json:"config_required"`
}

type ServiceCreateRequest struct {
	Name             string                   `json:"name"`
	Description      string                   `json:"description"`
	Properties       map[string]string        `json:"properties,omitempty"`
	ConfigParameters []ServiceConfigParameter `json:"config_required,omitempty"`
}

/*
Services Device Config Requests Look Like The Following:
[
  {
    "id": "592c8a627d6ec25f901d9687",
    "type": "device",
    "config": [
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

type ServiceConfigParameter struct {
	Name        string `json:"key_name"` // The key_ is redundant
	Description string `json:"key_description"`
	Example     string `json:"key_example"`
	Required    bool   `json:"key_required"`
}

// KeyValuePair represents the REST interface's internal structure for
// maps. This is typically just used to parse JSON from the REST interface.
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
func (n ServiceDeviceListItem) String() string {
	buf, _ := json.MarshalIndent(&n, "", jsonPrettyIndent)
	return string(buf)
}

func (n ServiceNode) String() string {
	buf, _ := json.MarshalIndent(&n, "", jsonPrettyIndent)
	return string(buf)
}
func (n ServiceCreateRequest) String() string {
	buf, _ := json.MarshalIndent(&n, "", jsonPrettyIndent)
	return string(buf)
}
func (n ServiceConfigParameter) String() string {
	buf, _ := json.MarshalIndent(&n, "", jsonPrettyIndent)
	return string(buf)
}
func (n KeyValuePair) String() string {
	buf, _ := json.MarshalIndent(&n, "", jsonPrettyIndent)
	return string(buf)
}

// RequestServiceInfo makes an HTTP GET to the framework server requesting
// the Service Node information for service with ID serviceid.
func (host Host) RequestServiceInfo(serviceid string) (ServiceNode, error) {
	var serviceNode ServiceNode
	uri := host.uri + rootAPISubPath + servicesSubPath + "/" + serviceid
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return serviceNode, err
	}
	req.SetBasicAuth(host.user, host.pass)

	// resp, err := http.Get(host.uri + servicesSubPath + "/" + serviceid)
	resp, err := host.client.Do(req)
	if err != nil {
		// should report auth problems here in future
		return serviceNode, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != httpStatusCodeOK {
		return serviceNode, fmt.Errorf(resp.Status)
	}
	err = json.NewDecoder(resp.Body).Decode(&serviceNode)
	return serviceNode, err
}

// RequestServiceDeviceList
func (host Host) RequestServiceDeviceList(serviceid string) ([]ServiceDeviceListItem, error) {
	var serviceDeviceListItems = make([]ServiceDeviceListItem, 0)
	uri := host.uri + rootAPISubPath + servicesSubPath + "/" + serviceid + serviceDevicesSubPath
	fmt.Println("Service URI: ", uri)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return serviceDeviceListItems, err
	}
	req.SetBasicAuth(host.user, host.pass)

	resp, err := host.client.Do(req)
	if err != nil {
		// should report auth problems here in future
		return serviceDeviceListItems, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != httpStatusCodeOK {
		return serviceDeviceListItems, fmt.Errorf(resp.Status)
	}
	err = json.NewDecoder(resp.Body).Decode(&serviceDeviceListItems)
	return serviceDeviceListItems, err
}
