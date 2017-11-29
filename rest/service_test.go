package rest_test

import (
	"os"
	"testing"

	"github.com/openchirp/framework/rest"
)

func TestHost_ServiceCreateAndCheck(t *testing.T) {
	// Get parameters from environment variables
	frameworkUri := os.Getenv("FRAMEWORK_SERVER")
	id := os.Getenv("USER_ID")
	token := os.Getenv("USER_TOKEN")

	host := rest.NewHost(frameworkUri)
	if err := host.Login(id, token); err != nil {
		t.Error("Error logging in:", err)
		return
	}

	sInfo1, err := host.ServiceCreate("Test Service 1", "My Test Service 1", nil, nil)
	if err != nil {
		t.Error("Error creating service:", err)
		return
	}
	t.Log(sInfo1)

	sInfo2, err := host.RequestServiceInfo(sInfo1.ID)
	if err != nil {
		t.Error("Error creating service:", err)
		return
	}
	t.Log(sInfo2)

	if sInfo1.ID != sInfo2.ID {
		t.Error("IDs don't match")
		return
	}

	if sInfo1.Name != sInfo2.Name {
		t.Error("Names don't match")
		return
	}

	if sInfo1.Description != sInfo2.Description {
		t.Error("Descriptions don't match")
		return
	}
}

func TestHost_ServiceCreateAndDelete(t *testing.T) {
	// Get parameters from environment variables
	frameworkUri := os.Getenv("FRAMEWORK_SERVER")
	id := os.Getenv("USER_ID")
	token := os.Getenv("USER_TOKEN")

	host := rest.NewHost(frameworkUri)
	if err := host.Login(id, token); err != nil {
		t.Error("Error logging in:", err)
		return
	}

	sInfo, err := host.ServiceCreate("Test Service 2", "My Test Service 2", nil, nil)
	if err != nil {
		t.Error("Error creating service:", err)
		return
	}
	t.Log(sInfo)

	err = host.ServiceDelete(sInfo.ID)
	if err != nil {
		t.Error("Error deleting service:", err)
		return
	}
}
