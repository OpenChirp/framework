package framework_test

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/openchirp/framework"
)

const (
	// The subscription key used to identify a messages types
	rawRxKey = 0
	rawTxKey = 1
)

// CDevice holds any data you want to keep around for a specific
// device that has linked your service.
//
// In this example, we will keep track of the rawrx and rawtx message counts
type CDevice struct {
	rawRxCount int
	rawTxCount int
}

// NewCDevice is called by the framework when a new device has been linked.
func NewCDevice() framework.Device {
	d := new(CDevice)
	// The following initialization is redundant in Go
	d.rawRxCount = 0
	d.rawTxCount = 0
	// Change type to the Device interface
	return framework.Device(d)
}

// ProcessLink is called once, during the initial setup of a
// device, and is provided the service config for the linking device.
func (d *CDevice) ProcessLink(ctrl *framework.DeviceControl) string {
	// Subscribe to subtopic "rawrx"
	ctrl.Subscribe("rawrx", rawRxKey)
	// Subscribe to subtopic "rawtx"
	ctrl.Subscribe("rawtx", rawTxKey)

	// This message is sent to the service status for the linking device
	return "Success"
}

// ProcessUnlink is called once, when the service has been unlinked from
// the device.
func (d *CDevice) ProcessUnlink(ctrl *framework.DeviceControl) {
	// The framework already handles unsubscribing from all
	// Device associted subtopics, so we don't need to call
	// ctrl.Unsubscribe.
}

// ProcessConfigChange is intended to handle a service config updates.
// If your program does not need to handle incremental config changes,
// simply return false, to indicate the config update was unhandled.
// The framework will then automatically issue a ProcessUnlink and then a
// ProcessLink, instead. Note, NewCDevice is not called.
//
// For more information about this or other Device interface functions,
// please see https://godoc.org/github.com/OpenChirp/framework#Device .
func (d *CDevice) ProcessConfigChange(ctrl *framework.DeviceControl, cchanges, coriginal map[string]string) (string, bool) {
	return "", false

	// If we have processed this config change, we should return the
	// new service status message and true.
	//
	//return "Sucessfully updated", true
}

// ProcessMessage is called upon receiving a pubsub message destined for
// this CDevice.
// Along with the standard DeviceControl object, the handler is provided
// a Message object, which contains the received message's payload,
// subtopic, and the provided Subscribe key.
func (d *CDevice) ProcessMessage(ctrl *framework.DeviceControl, msg framework.Message) {

	if msg.Key().(int) == rawRxKey {
		d.rawRxCount++
		subtopic := "rawrxcount"
		ctrl.Publish(subtopic, fmt.Sprint(d.rawRxCount))
	} else if msg.Key().(int) == rawTxKey {
		d.rawTxCount++
		subtopic := "rawtxcount"
		ctrl.Publish(subtopic, fmt.Sprint(d.rawTxCount))
	} else {
		log.Fatalln("Received unassociated message")
	}
}

func ExampleStartServiceClientManaged_counter() {
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
		NewCDevice)
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

	/* Setup signal channel */
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	/* Wait on a signal */
	<-signals
	log.Println("Shutting down")

	/* Post service's global status */
	err = c.SetStatus("Shutting down")
	if err != nil {
		log.Fatalln("Failed to publish service status: ", err)
		return
	}
	log.Println("Published Service Status")
}
