// December 29, 2017
// Craig Hesling <craig@hesling.com>
// TODO: Make user info send id instead of _id
// TODO: Make group info send id instead of group_id

package rest

import (
	"bytes"
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

// User is a container for summary User object received
// from the RESTful JSON interface
type User struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	UserID string `json:"userid"`
}

// UserDetails is a container for User info object received
// from the RESTful JSON interface
type UserDetails struct {
	User
	Groups []GroupNode `json:"groups"`
}

type UserCreateRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name,omitempty"`
	Password string `json:"password"`
}

func (n GroupNode) String() string {
	buf, _ := json.MarshalIndent(&n, "", jsonPrettyIndent)
	return string(buf)
}

func (n User) String() string {
	buf, _ := json.MarshalIndent(&n, "", jsonPrettyIndent)
	return string(buf)
}

func (n UserDetails) String() string {
	buf, _ := json.MarshalIndent(&n, "", jsonPrettyIndent)
	return string(buf)
}

// RequestUserInfo makes an HTTP GET to the framework server requesting
// the User Node information for user authenticated.
func (host Host) RequestUserInfo() (UserDetails, error) {
	var user UserDetails
	uri := host.uri + rootAPISubPath + userSubPath
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return user, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(host.user, host.pass)

	resp, err := host.client.Do(req)
	if err != nil {
		// should report auth problems here in future
		return user, err
	}
	defer resp.Body.Close()
	if err := DecodeOCError(resp); err != nil {
		return user, err
	}
	err = json.NewDecoder(resp.Body).Decode(&user)
	return user, err
}

// AllUsers makes an HTTP GET to the framework server requesting
// the all user summaries
func (host Host) AllUsers() ([]User, error) {
	var users []User
	uri := host.uri + rootAPISubPath + userSubPath + "/all"
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return users, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(host.user, host.pass)

	resp, err := host.client.Do(req)
	if err != nil {
		// should report auth problems here in future
		return users, err
	}
	defer resp.Body.Close()
	if err := DecodeOCError(resp); err != nil {
		return users, err
	}
	err = json.NewDecoder(resp.Body).Decode(&users)
	return users, err
}

// UserCreate requests the new user be created with the given
// name, email, and password
func (host Host) UserCreate(email, name, password string) error {
	uri := host.uri + authAPISubPath + "/signup"

	userReq := &UserCreateRequest{
		Email:    email,
		Name:     name,
		Password: password,
	}

	body, err := json.Marshal(userReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", uri, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := host.client.Do(req)
	if err != nil {
		// should report auth problems here in future
		return err
	}
	defer resp.Body.Close()
	return DecodeOCError(resp)
}
