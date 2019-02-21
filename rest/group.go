package rest

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// Group is a container for the minimal group summary
type Group struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GroupCreateRequest is the container for the request to create a new group
type GroupCreateRequest struct {
	Name string `json:"name"`
}

// GroupCreate requests for a new group to be created with the given
// name
func (host Host) GroupCreate(name string) error {
	uri := host.uri + rootAPISubPath + groupSubPath

	groupReq := &GroupCreateRequest{
		Name: name,
	}

	body, err := json.Marshal(groupReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", uri, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(host.user, host.pass)

	resp, err := host.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return DecodeOCError(resp)
}

// GroupAll fetches a list of all groups
func (host Host) GroupAll() ([]Group, error) {
	var groups []Group
	uri := host.uri + rootAPISubPath + groupSubPath
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return groups, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(host.user, host.pass)

	resp, err := host.client.Do(req)
	if err != nil {
		return groups, err
	}
	defer resp.Body.Close()
	if err := DecodeOCError(resp); err != nil {
		return groups, err
	}
	err = json.NewDecoder(resp.Body).Decode(&groups)
	return groups, err
}
