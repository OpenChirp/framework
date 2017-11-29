// December 29, 2017
// Craig Hesling <craig@hesling.com>
// TODO: Make user info send id instead of _id
// TODO: Make group info send id instead of group_id

package rest

import (
	"encoding/json"
	"net/http"
)

// GroupNode is a container for User Group object received
// from the RESTful JSON interface
type GroupNode struct {
	ID          string `json:"group_id"` // Should really just be id
	Name        string `json:"name"`
	WriteAccess bool   `json:"write_access"`
}

// UserNode is a container for User Node object received
// from the RESTful JSON interface
type UserNode struct {
	// We currently omit the _id
	// TODO: Rename _id to id in REST interface
	Name   string      `json:"name"`
	Email  string      `json:"email"`
	UserID string      `json:"userid"`
	Groups []GroupNode `json:"groups"`
}

func (n GroupNode) String() string {
	buf, _ := json.MarshalIndent(&n, "", jsonPrettyIndent)
	return string(buf)
}

func (n UserNode) String() string {
	buf, _ := json.MarshalIndent(&n, "", jsonPrettyIndent)
	return string(buf)
}

// RequestUserInfo makes an HTTP GET to the framework server requesting
// the User Node information for user authenticated.
func (host Host) RequestUserInfo() (UserNode, error) {
	var userNode UserNode
	uri := host.uri + rootAPISubPath + userSubPath
	req, err := http.NewRequest("GET", uri, nil)
	req.SetBasicAuth(host.user, host.pass)

	resp, err := host.client.Do(req)
	if err != nil {
		// should report auth problems here in future
		return userNode, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&userNode)
	return userNode, err
}
