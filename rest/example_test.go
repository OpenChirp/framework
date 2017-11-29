package rest_test

import (
	"fmt"
	"log"
	"os"

	"github.com/openchirp/framework/rest"
)

func ExampleNewHost() {
	frameworkUri := "http://localhost:7000"
	id := "5a1ea73df76abe01c57abfb8"
	token := "DJpHxwmExGbcYwsEHgQezDVeKS4N"

	host := rest.NewHost(frameworkUri)
	if err := host.Login(id, token); err != nil {
		log.Fatalln("Error logging in:", err)
	}
	log.Println(host)
}

func ExampleHost_RequestServiceInfo() {
	// Get parameters from environment variables
	frameworkUri := os.Getenv("FRAMEWORK_SERVER")
	id := os.Getenv("SERVICE_ID")
	token := os.Getenv("SERVICE_TOKEN")

	host := rest.NewHost(frameworkUri)
	if err := host.Login(id, token); err != nil {
		log.Fatalln("Error logging in:", err)
	}

	sInfo, err := host.RequestServiceInfo(id)
	if err != nil {
		log.Fatalln("Error requesting service info:", err)
	}
	fmt.Println(sInfo)
	// Ouput: Blah
}

func ExampleHost_RequestUserInfo() {
	// Get parameters from environment variables
	frameworkUri := os.Getenv("FRAMEWORK_SERVER")
	id := os.Getenv("USER_ID")
	token := os.Getenv("USER_TOKEN")

	host := rest.NewHost(frameworkUri)
	if err := host.Login(id, token); err != nil {
		log.Fatalln("Error logging in:", err)
	}

	uInfo, err := host.RequestUserInfo()
	if err != nil {
		log.Fatalln("Error requesting user info:", err)
	}
	fmt.Println(uInfo)
}
