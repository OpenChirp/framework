// TODO: Handle errors for pubsub methods, although these errors are probably
// fatal.
// TODO: Straighten out the case for when a user subscribes
//       to the same topic twice.

package framework

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/golang/groupcache/lru"
)

const (
	deviceCtrlsCacheSize = 100
)

type serviceManager struct {
	c           *ServiceClient
	newdevice   func() Device
	updates     <-chan DeviceUpdate
	devices     map[string]*deviceState
	deviceCtrls *lru.Cache
	shutdown    chan bool
	wg          sync.WaitGroup
}

// runtime is the primary service manager routine that handles device service
// config and link changes
func (m *serviceManager) runtime() {
	defer m.wg.Done()

	for {
		select {
		case update := <-m.updates:
			switch update.Type {
			case DeviceUpdateTypeRem:
				m.removeDevice(update.Id)
			case DeviceUpdateTypeUpd:
				fallthrough
			case DeviceUpdateTypeAdd:
				m.addUpdateDevice(update.Id, update.Topic, update.Config)
			}
		case <-m.shutdown:
			return
		}
	}
}

func (m *serviceManager) Stop() {
	m.shutdown <- true
	m.wg.Wait()
	m.c.manager = nil
}

/* DeviceControl Cache */

func (m *serviceManager) deviceCtrlsCacheAdd(dCtrl *DeviceControl) {
	m.deviceCtrls.Add(lru.Key(dCtrl.Id()), dCtrl)
}

func (m *serviceManager) deviceCtrlsCacheGet(deviceid string) (*DeviceControl, bool) {
	dCtrlInt, dCtrlExists := m.deviceCtrls.Get(lru.Key(deviceid))
	if dCtrlExists {
		if dCtrl, ok := dCtrlInt.(*DeviceControl); ok {
			// TODO: Should probably assert that the dCtrl.dStat == m.devices[deviceid]
			return dCtrl, true
		}
	}
	return nil, false
}

func (m *serviceManager) deviceCtrlsCacheRemove(deviceid string) {
	m.deviceCtrls.Remove(lru.Key(deviceid))
}

func (m *serviceManager) deviceCtrlsCacheProvide(dState *deviceState) *DeviceControl {
	dCtrlInt, dCtrlExists := m.deviceCtrls.Get(lru.Key(dState.id))
	if !dCtrlExists {
		dCtrlInt = m.generateDeviceCtrl(dState)
		m.deviceCtrlsCacheAdd(dCtrlInt.(*DeviceControl))
	}
	// TODO: Should probably assert that the dCtrl.dStat == m.devices[deviceid]
	return dCtrlInt.(*DeviceControl)
}

/* Service Manager Event Functions */

func (m *serviceManager) addUpdateDevice(deviceid string, topic string, config map[string]string) {
	if dState, dStateExists := m.devices[deviceid]; dStateExists {
		// Find config differences
		cchanges, missingKeys := configChanges(dState.config, config)
		if missingKeys {
			// Do not allow keys to be missing, since we do not expect users to
			// to understand missing keys on updates - we will remove and re-add
			// TODO: Should probably log, since this may be a REST bug
			log.Printf("missing keys, but the changes were: %v", cchanges)
			m.removeDevice(deviceid)
			m.addUpdateDevice(deviceid, topic, config)
			return
		}

		// Check that key/values actually changed
		if len(cchanges) == 0 {
			// If no changes are necessary, we are done
			return
		}

		// Save original config
		coriginal := dState.config

		// Set new config
		dState.config = config

		// Fetch a device control
		dCtrl := m.deviceCtrlsCacheProvide(dState)

		// Allow service to handle incremental config change
		status, ack := dState.userDevice.ProcessConfigChange(dCtrl, cchanges, coriginal)
		if !ack {
			// If the user refused to acknowledge a config update - we will
			// remove and re-add the link
			// 1. Restore original config
			dState.config = coriginal
			// 2. Run through removal process
			m.removeDevice(deviceid)
			// 3. Set new config
			dState.config = config
			// 4. Run through add process
			m.addUpdateDevice(deviceid, topic, config)
			return
		}

		// Update device's service link status
		m.c.SetDeviceStatus(dState.id, status)
	} else {
		// Create a new device context
		dState := &deviceState{
			id:         deviceid,
			topic:      topic,
			config:     config,
			subs:       make(map[string]interface{}),
			userDevice: m.newdevice(),
		}
		m.devices[deviceid] = dState

		// Fetch a device control
		dCtrl := m.deviceCtrlsCacheProvide(dState)

		// Process link
		status := dState.userDevice.ProcessLink(dCtrl)

		// Update device's service link status
		m.c.SetDeviceStatus(dState.id, status)
	}

}

func (m *serviceManager) removeDevice(deviceid string) {
	if dState, dStateExists := m.devices[deviceid]; dStateExists {
		// Fecth a device control
		dCtrl := m.deviceCtrlsCacheProvide(dState)

		// Process unlink
		dState.userDevice.ProcessUnlink(dCtrl)

		// Unsubscribe from all remaining topics
		m.deviceUnsubscribeAll(dState)

		// Delete device context
		delete(m.devices, deviceid)

		// We must remove dCtrl from cache, since we will be creating a new
		// deviceState.
		m.deviceCtrlsCacheRemove(deviceid)
	}
}

func (m *serviceManager) generateDeviceCtrl(dState *deviceState) *DeviceControl {
	return &DeviceControl{
		manager: m,
		dState:  dState,
	}
}

// deviceUnsubscribe unsubscribes from topics within the device's subtopic space
func (m *serviceManager) deviceUnsubscribeAll(dState *deviceState) {
	// Create a flat array of device subscribed topics
	topics := make([]string, 0, len(dState.subs))
	for topic := range dState.subs {
		topics = append(topics, topic)
	}
	// Unsubscribe from all device subscribed topics
	m.c.Unsubscribe(topics...)
	// Reset device's subscription list
	dState.subs = make(map[string]interface{})
}

// deviceUnsubscribe unsubscribes from topics within the device's subtopic space
func (m *serviceManager) deviceUnsubscribe(dState *deviceState, subtopics ...string) {
	// Prepend the device endpoint and remove from device subscription list
	for i, subtopic := range subtopics {
		topic := dState.topic + "/" + subtopic
		subtopics[i] = topic
		delete(dState.subs, topic)
	}
	// Unsubscribe from specified topics
	m.c.Unsubscribe(subtopics...)
}

// deviceSubscribe subscribes to a topic within the device's subtopic space.
// Currently, only the first call to subscribe to a particular topic is used.
//
// Messages received on the subscribed topic will be sent to the device's
// ProcessMessage handler with the specified key and subtopic.
func (m *serviceManager) deviceSubscribe(dState *deviceState, subtopic string, key interface{}) {
	stopic := dState.topic + "/" + subtopic
	if _, ok := dState.subs[stopic]; !ok {
		m.c.Subscribe(stopic, func(topic string, payload []byte) {
			// Get the device level subtopic
			subtopic := strings.TrimPrefix(topic, dState.topic)
			// Compose message for device message handler
			msg := Message{
				key:     key,
				topic:   subtopic,
				payload: payload,
			}
			// Fetch a device control object device message handler
			dCtrl := m.deviceCtrlsCacheProvide(dState)
			// Run device message handler
			dState.userDevice.ProcessMessage(dCtrl, msg)
		})
		dState.subs[stopic] = key
	}
}

// devicePublish publishes to a topic within the device's subtopic space
func (m *serviceManager) devicePublish(dState *deviceState, subtopic string, payload interface{}) {
	topic := dState.topic + "/" + subtopic
	m.c.Publish(topic, payload)
}

type deviceState struct {
	userDevice Device
	id         string
	topic      string
	config     map[string]string
	subs       map[string]interface{}
}

// StartServiceClientManaged starts the service client layer using the fully
// managed mode
func StartServiceClientManaged(
	frameworkuri,
	brokeruri,
	id,
	token,
	statusmsg string,
	newdevice func() Device,
) (*ServiceClient, error) {

	if newdevice == nil {
		return nil, fmt.Errorf("Error: newdevice cannot be nil")
	}

	c, err := StartServiceClientStatus(frameworkuri, brokeruri, id, token, statusmsg)
	if err != nil {
		return nil, err
	}

	manager := new(serviceManager)
	manager.c = c
	manager.newdevice = newdevice
	manager.devices = make(map[string]*deviceState)
	manager.shutdown = make(chan bool)

	manager.deviceCtrls = lru.New(deviceCtrlsCacheSize)

	updates, err := c.StartDeviceUpdatesSimple()
	if err != nil {
		c.StopClient()
		return nil, err
	}
	manager.updates = updates

	manager.wg.Add(1)
	c.manager = manager
	go manager.runtime()

	return c, nil
}

// Device is the interface services will implement
type Device interface {
	// ProcessLink is called once, during the initial setup of a
	// device, and is provided the service config for the linking device.
	// The service is expected to parse the provided config for initial setup.
	// The returned string is used as the device's link status.
	ProcessLink(ctrl *DeviceControl) string
	// ProcessUnlink is called once, when the service has been unlinked from
	// the device.
	ProcessUnlink(ctrl *DeviceControl)
	// ProcessConfigChange is called only when the config has truly changed.
	// The specific config key/values which changed are provided in cchange
	// and the original config is provided in coriginal. Upon successful
	// completion of this call, the device's link status will be updated with
	// the returned string.
	//
	// If you do not want to handle incremental config changes, you may return
	// false. In this case, the service manager will restore the original
	// config, call for the device to be unlinked(ProcessUnlinked), clear the
	// device context, and call ProcessLink with the new config.
	// Note that the new config is accessible through ctrl.Config()
	ProcessConfigChange(ctrl *DeviceControl, cchanges, coriginal map[string]string) (string, bool)
	// ProcessMessage is called upon receiving a pubsub message destined for
	// this device. Along with the standard DeviceControl object, the
	// handler is provided a Message object, which contains the received
	// message's payload, subtopic, and the provided Subscribe key.
	ProcessMessage(ctrl *DeviceControl, msg Message)
}

// DeviceControl provides a simplified set of methods for controlling
// a single device. A DeviceContol object is provided within the context
// of a single device.
// The key uses are to Subscribe/Publish/Unsubscribe (pubsub methods)
// to a device's subtopic and to present the device's current Config and Id.
// Understand that the pubsub methods will automatically prepend device's
// topic prefix (ex. openchirp/device/<device_id>/) to the specified subtopic.
//
// Additionally, you should note that the Pubsub methods do not return errors
// and do not ask you to provide message handler functions.
// This shifts the responsibility of error handling and message passing
// to the Managed Service client.
type DeviceControl struct {
	manager *serviceManager
	dState  *deviceState
}

// Id returns this device's id
func (c *DeviceControl) Id() string {
	return c.dState.id
}

// Config returns this device's current config
func (c *DeviceControl) Config() map[string]string {
	return c.dState.config
}

// Subscribe to a device's subtopic and associate with key.
//
// When receiving a message for this subtopic, the Device's
// ProcessMessage handler will be invoked with the message
// and the this key. See
func (c *DeviceControl) Subscribe(subtopic string, key interface{}) {
	c.manager.deviceSubscribe(c.dState, subtopic, key)
}

// Unsubscribe unsubscribes from the specified device's subtopics
func (c *DeviceControl) Unsubscribe(subtopics ...string) {
	c.manager.deviceUnsubscribe(c.dState, subtopics...)
}

// UnsubscribeAll unsubscribes from all of this device's subtopics
func (c *DeviceControl) UnsubscribeAll() {
	c.manager.deviceUnsubscribeAll(c.dState)
}

// Publish publishes payload to this device's subtopic
func (c *DeviceControl) Publish(subtopic string, payload interface{}) {
	c.manager.devicePublish(c.dState, subtopic, payload)
}

// Message holds a received pubsub payload and topic along with the
// provided subscription key
type Message struct {
	key     interface{}
	topic   string
	payload []byte
}

// String shows all parts of the message as a human readable string
func (t Message) String() string {
	return fmt.Sprintf("%v: %s: [ % #xv ]", t.key, t.topic, t.payload)
}

// Key returns the provided subscription key for this message
func (t Message) Key() interface{} {
	return t.key
}

// Topic returns the pubsub subtopic which received this message
func (t Message) Topic() interface{} {
	return t.topic
}

// Payload returns the pubsub payload of this message
func (t Message) Payload() []byte {
	return t.payload
}

// configChanges returns a map of only the keys that changed.
// If keys were deleted from the newer config, the return bool will be true.
func configChanges(original, new map[string]string) (map[string]string, bool) {
	var omittedKey bool
	m := make(map[string]string, len(new))

	// Make copy of new config
	for k, v := range new {
		m[k] = v
	}

	// Remove keys that didn't change - keep track of how many
	for k, v := range original {
		if nv, ok := m[k]; ok && (v == nv) {
			delete(m, k)
		} else if !ok {
			// when a key is missing from the new config, we assign it ""
			omittedKey = true
			m[k] = ""
		}
	}
	return m, omittedKey
}
