package framework

import "encoding/json"

type PubSub struct {
	Protocol  string          `json:"protocol"`
	Topic     string          `json:"endpoint"`
	ExtraJunk json.RawMessage `json:"serviceconfig"`
}
