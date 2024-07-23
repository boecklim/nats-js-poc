package common

import (
	"context"
	"encoding/json"
	"fmt"

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
	JetStream    jetstream.JetStream
	JetStreamCfg jetstream.StreamConfig
	nc           *nats.Conn
}

func NewJetStreamClient(url string) (*JetStreamClient, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}

	cfg := jetstream.StreamConfig{
		Name:      StreamName,
		Retention: jetstream.WorkQueuePolicy,
		Subjects:  []string{Subject},
		Storage:   jetstream.MemoryStorage,
	}

	p := &JetStreamClient{
		JetStream:    js,
		JetStreamCfg: cfg,
		nc:           nc,
	}

	return p, nil
}

func (js *JetStreamClient) Close() error {
	return js.nc.Drain()
}

func PrintStreamState(ctx context.Context, stream jetstream.Stream) error {
	info, err := stream.Info(ctx)
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(info.State, "", " ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))

	return nil
}
