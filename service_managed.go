package framework

import (
	"fmt"
	"sync"

	"github.com/golang/groupcache/lru"
)

const (
	deviceCtrlsCacheSize = 100
	devicePrefix         = "openchirp/devices"
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
				m.addUpdateDevice(update.Id, update.Config)
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

func (m *serviceManager) deviceCtrlsCacheAdd(deviceid string, dCtrl *DeviceControl) {
	m.deviceCtrls.Add(lru.Key(deviceid), dCtrl)
}

func (m *serviceManager) addUpdateDevice(deviceid string, config map[string]string) {
	dState, dStateExists := m.devices[deviceid]
	dCtrl, dCtrlExists := m.deviceCtrlsCacheGet(deviceid)
	if dStateExists {
		coriginal := dState.config
		cchanges, missingKeys := configChanges(coriginal, config)
		if missingKeys {
			// Do not allow keys to be missing, since we do not expect users to
			// to understand missing keys on updates - we will remove and re-add
			// TODO: Should probably log, since this may be a REST bug
			m.removeDevice(deviceid)
			m.addUpdateDevice(deviceid, config)
			return
		}
		dState.config = config

		if !dCtrlExists {
			dCtrl = m.generateDeviceCtrl(dState)
			m.deviceCtrlsCacheAdd(deviceid, dCtrl)
		}

		if !dState.userDevice.ProcessConfigChange(dCtrl, cchanges, coriginal) {
			// If the user refused to acknowledge a config update - we will
			// remove and re-add
			m.removeDevice(deviceid)
			m.addUpdateDevice(deviceid, config)
			return
		}
	} else {
		dState := &deviceState{
			id:         deviceid,
			config:     config,
			subs:       make(map[string]interface{}),
			userDevice: m.newdevice(),
		}
		m.devices[deviceid] = dState
		dCtrl := m.generateDeviceCtrl(dState)
		m.deviceCtrlsCacheAdd(deviceid, dCtrl)
		dState.userDevice.ProcessLink(dCtrl)
	}
}

func (m *serviceManager) removeDevice(deviceid string) {
	dState, dStateExists := m.devices[deviceid]
	dCtrl, dCtrlExists := m.deviceCtrlsCacheGet(deviceid)
	if dStateExists {
		if !dCtrlExists {
			dCtrl = m.generateDeviceCtrl(dState)
			m.deviceCtrlsCacheAdd(deviceid, dCtrl)
		}
		dState.userDevice.ProcessUnlink(dCtrl)
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

func (m *serviceManager) deviceUnsubscribeAll(dState *deviceState) {
	for topic, _ := range dState.subs {
		m.c.Unsubscribe(topic)
		delete(dState.subs, topic)
	}
}

func (m *serviceManager) deviceUnsubscribe(dState *deviceState, subtopics ...string) {
	for _, subtopic := range subtopics {
		topic := devicePrefix + dState.id + "/" + subtopic
		if _, ok := dState.subs[topic]; ok {
			m.c.Unsubscribe(topic)
			delete(dState.subs, topic)
		}
	}
}

func (m *serviceManager) deviceSubscribe(dState *deviceState, subtopic string, key interface{}) {
	stopic := devicePrefix + dState.id + "/" + subtopic
	if _, ok := dState.subs[stopic]; !ok {
		m.c.Subscribe(stopic, func(topic string, payload []byte) {
			msg := Message{
				key:     key,
				topic:   topic,
				payload: payload,
			}
			dCtrl, dCtrlExists := m.deviceCtrlsCacheGet(dState.id)
			if !dCtrlExists {
				dCtrl = m.generateDeviceCtrl(dState)
				m.deviceCtrlsCacheAdd(dState.id, dCtrl)
			}
			dState.userDevice.ProcessMessage(dCtrl, msg)
		})
		dState.subs[stopic] = key
	}
}

func (m *serviceManager) devicePublish(dState *deviceState, subtopic string, payload interface{}) {
	topic := devicePrefix + dState.id + "/" + subtopic
	m.c.Publish(topic, payload)
}

type deviceState struct {
	userDevice Device
	id         string
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

type Device interface {
	ProcessLink(ctrl *DeviceControl)
	ProcessUnlink(ctrl *DeviceControl)
	ProcessConfigChange(ctrl *DeviceControl, cchanges, coriginal map[string]string) bool
	ProcessMessage(ctrl *DeviceControl, msg Message)
}

// DeviceControl aims to provide the Device implementation an error
// free set of pubsub methods which are scoped to the OC Device's
// pubsub prefix. Additionally, the subscribe method does not ask for
// a callback function, since it is the responsibility of the Service client
// to provide received messages to the Device implementation.
type DeviceControl struct {
	manager *serviceManager
	dState  *deviceState
}

func (c *DeviceControl) Id() string {
	return c.dState.id
}

func (c *DeviceControl) Config() map[string]string {
	return c.dState.config
}

func (c *DeviceControl) Subscribe(subtopic string, key interface{}) {
	c.manager.deviceSubscribe(c.dState, subtopic, key)
}

func (c *DeviceControl) Unsubscribe(subtopics ...string) {
	c.manager.deviceUnsubscribe(c.dState, subtopics...)
}

func (c *DeviceControl) Publish(subtopic string, payload []interface{}) {
	c.manager.devicePublish(c.dState, subtopic, payload)
}

type Message struct {
	key     interface{}
	topic   string
	payload []byte
}

func (t Message) String() string {
	return fmt.Sprintf("%v: %s: [ % #xv ]", t.key, t.topic, t.payload)
}

func (t Message) Key() interface{} {
	return t.Key
}

func (t Message) Topic() interface{} {
	return t.Topic
}

func (t Message) Payload() interface{} {
	return t.payload
}

// Returns a map of only the keys that changed.
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
		} else if ok {
			// when a key is missing from the new config, we assign it ""
			omittedKey = true
			m[k] = ""
		}
	}
	return m, omittedKey
}
