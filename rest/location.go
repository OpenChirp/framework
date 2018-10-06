// December 29, 2017
// Craig Hesling <craig@hesling.com>

package rest

import (
	"encoding/json"
	"net/http"
)

// LocationNode is a container for Location Node object received
// from the RESTful JSON interface
type LocationNode struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Owner    Owner    `json:"owner"`
	Children []string `json:"children"`
	// We currently omit the geo_loc, timestamps, test, and type
}

func (n LocationNode) String() string {
	buf, _ := json.MarshalIndent(&n, "", jsonPrettyIndent)
	return string(buf)
}

// RequestLocationInfo makes an HTTP GET to the framework server requesting
// the Location Node information for the location with ID locid.
func (host Host) RequestLocationInfo(locID string) (LocationNode, error) {
	var locNode LocationNode
	var uri string
	if locID == "" {
		uri = host.uri + rootAPISubPath + locationSubPath
	} else {
		uri = host.uri + rootAPISubPath + locationSubPath + "/" + locID
	}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return locNode, err
	}
	req.SetBasicAuth(host.user, host.pass)

	resp, err := host.client.Do(req)
	if err != nil {
		// should report auth problems here in future
		return locNode, err
	}
	defer resp.Body.Close()
	if locID == "" {
		// TODO: Figure out why the root node is in an array
		var roots []LocationNode
		err = json.NewDecoder(resp.Body).Decode(&roots)
		if err != nil {
			return locNode, err
		}
		if len(roots) < 1 {
			return locNode, err
		}
		locNode = roots[0]
	} else {
		err = json.NewDecoder(resp.Body).Decode(&locNode)
	}

	return locNode, err
}

// RequestLocationDevices makes an HTTP GET to the framework server requesting
// the the list of devices at the specified location.
// If recursive is true, devices that are located on any sublocations will be
// included.
func (host Host) RequestLocationDevices(locID string, recursive bool) ([]NodeDescriptor, error) {
	var deviceNodes []NodeDescriptor
	var uri string

	deviceSuffix := "/devices"
	if recursive {
		deviceSuffix = "/alldevices"
	}

	// Unfortunately, the location api doesn't allow location/devices for root
	if locID == "" {
		loc, err := host.RequestLocationInfo("")
		if err != nil {
			return nil, err
		}
		locID = loc.ID
	}

	uri = host.uri + rootAPISubPath + locationSubPath + "/" + locID + deviceSuffix
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return deviceNodes, err
	}
	req.SetBasicAuth(host.user, host.pass)

	resp, err := host.client.Do(req)
	if err != nil {
		// should report auth problems here in future
		return deviceNodes, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&deviceNodes)
	return deviceNodes, err
}
