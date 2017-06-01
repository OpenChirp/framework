package framework

/*
News Updates Look Like The Following:
openchirp/services/592880c57d6ec25f901d9668/thing/new
{
	"thing":{
		"type":"device",
		"id":"592c8a627d6ec25f901d9687",
		"config":[{"key":"DevEUI","value":"test1"},
					{"key":"AppEUI","value":"test2"},
					{"key":"AppKey","value":"test3"}]
		}
}
*/

type KeyValuePair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ServiceUpdatesEncapsulation struct {
	Thing ServiceDeviceUpdate `json:"thing"`
}

type ServiceDeviceUpdate struct {
	Id     string         `json:"id"`
	Config []KeyValuePair `json:"config"`
}
