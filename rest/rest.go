// December 14, 2016
// Craig Hesling <craig@hesling.com>

// Package rest provides the data structures and primitive mechanisms
// for representing and communicating framework constructs with the RESTful
// server.
package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	rootAPISubPath        = "/apiv1"
	authAPISubPath        = "/authv1"
	deviceSubPath         = "/device"
	servicesSubPath       = "/service"
	serviceDevicesSubPath = "/things"
	serviceTokenSubPath   = "/token"
	locationSubPath       = "/location"
	userSubPath           = "/user"
	groupSubPath          = "/group"
	healthCheckSubPath    = "/check"
)

const (
	httpStatusCodeOK = 200
)

const jsonPrettyIndent = "  "

func DecodeOCError(resp *http.Response) error {
	if resp == nil {
		return fmt.Errorf("Filed to decode response. Check err returned by http request.")
	}
	if resp.StatusCode == httpStatusCodeOK {
		return nil
	}
	var ocerror struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ocerror); err != nil {
		// failed to decode a message error
		// return server error message instead
		return fmt.Errorf(resp.Status)
	}

	// Insert blank error message
	if ocerror.Error.Message == "" {
		ocerror.Error.Message = "<Blank-Error-Message-Received>"
	}
	return fmt.Errorf(ocerror.Error.Message)
}

// Host represents the RESTful HTTP server that hosts the framework
type Host struct {
	uri string
	// This is where we add APIKeys and username/password for user
	user   string
	pass   string
	client http.Client
}

// NewHost returns an object referencing the framework server
func NewHost(uri string) Host {
	// no need to decompose uri using net/url package
	return Host{uri: uri, client: http.Client{}}
}

func (host *Host) Login(username, password string) error {
	host.user = username
	host.pass = password
	// TODO: Check login credentials -- return error if no good
	return nil
}

// PubSub describes a node's pubsub endpoint
type PubSub struct {
	Protocol string `json:"protocol"`
	Topic    string `json:"endpoint"`
}

// Owner describes the owning user's details
type Owner struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// NodeDescriptor provides the common fields that Device and Service nodes share
type NodeDescriptor struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Pubsub PubSub `json:"pubsub"`
	Owner  Owner  `json:"owner"`
}
