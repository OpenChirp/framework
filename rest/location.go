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
func (host Host) RequestLocationInfo(locid string) (LocationNode, error) {
	var locNode LocationNode
	var uri string
	if locid == "" {
		uri = host.uri + rootAPISubPath + locationSubPath
	} else {
		uri = host.uri + rootAPISubPath + locationSubPath + "/" + locid
	}
	req, err := http.NewRequest("GET", uri, nil)
	req.SetBasicAuth(host.user, host.pass)

	resp, err := host.client.Do(req)
	if err != nil {
		// should report auth problems here in future
		return locNode, err
	}
	defer resp.Body.Close()
	if locid == "" {
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
