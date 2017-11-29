package rest_test

import (
	"os"
	"testing"

	"github.com/openchirp/framework/rest"
)

func TestHost_RequestLocationInfo(t *testing.T) {
	// Get parameters from environment variables
	frameworkUri := os.Getenv("FRAMEWORK_SERVER")
	id := os.Getenv("USER_ID")
	token := os.Getenv("USER_TOKEN")

	host := rest.NewHost(frameworkUri)
	if err := host.Login(id, token); err != nil {
		t.Error("Error logging in:", err)
		return
	}

	lInfo, err := host.RequestLocationInfo("")
	if err != nil {
		t.Error("Error requesting root location info:", err)
		return
	}
	t.Log(lInfo)

	lInfo, err = host.RequestLocationInfo(lInfo.Children[0])
	if err != nil {
		t.Error("Error requesting location info:", err)
		return
	}
	t.Log(lInfo)
}
func TestHost_RequestUserInfo(t *testing.T) {
	// Get parameters from environment variables
	frameworkUri := os.Getenv("FRAMEWORK_SERVER")
	id := os.Getenv("USER_ID")
	token := os.Getenv("USER_TOKEN")

	host := rest.NewHost(frameworkUri)
	if err := host.Login(id, token); err != nil {
		t.Error("Error logging in:", err)
		return
	}

	uInfo, err := host.RequestUserInfo()
	if err != nil {
		t.Error("Error requesting user info:", err)
		return
	}
	t.Log(uInfo)
}
