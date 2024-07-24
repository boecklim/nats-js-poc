package common

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const StreamName = "stream-1"
const Subject = "submit-tx"

type Msg struct {
	Msg  string `json:"name"`
	UUID string `json:"uuid"`
}

type Client struct {
	JetStream    jetstream.JetStream
	JetStreamCfg jetstream.StreamConfig
	nc           *nats.Conn
	Logger       *slog.Logger
}

func NewJetStreamClient(url string, logger *slog.Logger, connectionName string) (*Client, error) {
	nc, err := nats.Connect(url,
		nats.Name(connectionName),
	)
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

	p := &Client{
		JetStream:    js,
		JetStreamCfg: cfg,
		nc:           nc,
		Logger:       logger,
	}

	return p, nil
}

func (js *Client) Close() error {
	return js.nc.Drain()
}

func (js *Client) GetStream(ctx context.Context) (jetstream.Stream, error) {
	stream, err := js.JetStream.Stream(ctx, StreamName)
	if errors.Is(err, jetstream.ErrStreamNotFound) {
		js.Logger.Error(fmt.Sprintf("stream %s not found, creating new", StreamName))

		stream, err = js.JetStream.CreateStream(ctx, js.JetStreamCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create stream: %v", err)
		}

		js.Logger.Info(fmt.Sprintf("stream %s created", StreamName))
		return stream, nil
	}

	js.Logger.Info(fmt.Sprintf("stream %s found", StreamName))
	return stream, nil
}
