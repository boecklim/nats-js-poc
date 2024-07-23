package common

import (
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const StreamName = "stream-1"
const Subject = "submit-tx"

type Msg struct {
	Msg string `json:"name"`
	ID  int    `json:"id"`
}

type JetStreamClient struct {
	Js  jetstream.JetStream
	Cfg jetstream.StreamConfig
}

func NewJetStreamClient(url string) (*JetStreamClient, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	defer nc.Drain()

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}

	cfg := jetstream.StreamConfig{
		Name:      StreamName,
		Retention: jetstream.WorkQueuePolicy,
		Subjects:  []string{},
		Storage:   jetstream.MemoryStorage,
	}

	p := &JetStreamClient{
		Js:  js,
		Cfg: cfg,
	}

	return p, nil
}
