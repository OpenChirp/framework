package framework_test

import (
	"log"

	"github.com/openchirp/framework"
)

type Device struct {
}

func NewDevice() framework.Device {
	d := new(Device)
	return framework.Device(d)
}
func (d *Device) ProcessLink(ctrl *framework.DeviceControl) string {
	return "Success"
}
func (d *Device) ProcessUnlink(ctrl *framework.DeviceControl) {
}
func (d *Device) ProcessConfigChange(ctrl *framework.DeviceControl, cchanges, coriginal map[string]string) (string, bool) {
	return "", false
}
func (d *Device) ProcessMessage(ctrl *framework.DeviceControl, msg framework.Message) {
}

// ExampleStartServiceClientManaged_minimal demonstates the minimal configuration
// to use StartServiceClientManaged
func ExampleStartServiceClientManaged_minimal() {
	// Parse parameters from command line or environment variables
	frameworkServer := "http://localhost:7000"
	mqttServer := "localhost:1883"
	serviceId := "5a1ea73df76abe01c57abfb8"
	serviceToken := "DJpHxwmExGbcYwsEHgQezDVeKS4N"

	c, err := framework.StartServiceClientManaged(
		frameworkServer,
		mqttServer,
		serviceId,
		serviceToken,
		"Unexpected disconnect!",
		NewDevice)
	if err != nil {
		log.Fatalln("Failed to StartServiceClient: ", err)
	}
	defer c.StopClient()
	log.Println("Started service")

	/* Post service's global status */
	err = c.SetStatus("Started")
	if err != nil {
		log.Fatalln("Failed to publish service status: ", err)
		return
	}
	log.Println("Published Service Status")
}
