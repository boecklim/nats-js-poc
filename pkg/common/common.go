package common

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const StreamName = "test-stream-1"
const StreamDesc = "Stream for testing JetStream as message queue"
const Subject = "test-subject"

type Msg struct {
	Msg  string `json:"name"`
	UUID string `json:"uuid"`
}

type Client struct {
	JetStream    jetstream.JetStream
	JetStreamCfg jetstream.StreamConfig
	nc           *nats.Conn
	Logger       *slog.Logger
	url          string
}

func NewJetStreamClient(url string, logger *slog.Logger) (*Client, error) {

	cfg := jetstream.StreamConfig{
		Name:        StreamName,
		Description: StreamDesc,
		Subjects:    []string{Subject},
		Retention:   jetstream.WorkQueuePolicy,
		// MaxConsumers: 0,
		// MaxMsgs:      0,
		// MaxBytes: 5000000000, // 5 GB
		MaxBytes: 50000, // 5 GB
		Discard:  jetstream.DiscardOld,
		// DiscardNewPerSubject: false,
		MaxAge: 10 * time.Minute,
		// MaxMsgsPerSubject:    0,
		// MaxMsgSize:           0,
		Storage: jetstream.MemoryStorage,
		// Replicas:             0,
		NoAck: false,
		// Duplicates: 0,
		// Placement:            &jetstream.Placement{},
		// Mirror:               &jetstream.StreamSource{},
		// Sources:              []*jetstream.StreamSource{},
		// Sealed:           false,
		// DenyDelete:       false,
		// DenyPurge:        false,
		// AllowRollup:      false,
		// Compression:      0,
		// FirstSeq:         0,
		// SubjectTransform: &jetstream.SubjectTransformConfig{},
		// RePublish:        &jetstream.RePublish{},
		// AllowDirect:      false,
		// MirrorDirect:     false,
		// ConsumerLimits:   jetstream.StreamConsumerLimits{},
		// Metadata:         map[string]string{},
		// Template:         "",
	}

	p := &Client{
		JetStreamCfg: cfg,
		Logger:       logger,
		url:          url,
	}

	return p, nil
}

func (cl *Client) Connect() error {

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	nc, err := nats.Connect(cl.url,
		nats.Name(hostname),
		nats.ErrorHandler(func(c *nats.Conn, s *nats.Subscription, err error) {
			cl.Logger.Error("connection error", slog.String("err", err.Error()))
		}),
		nats.DiscoveredServersHandler(func(nc *nats.Conn) {
			cl.Logger.Info(fmt.Sprintf("Known servers: %v\n", nc.Servers()))
			cl.Logger.Info(fmt.Sprintf("Discovered servers: %v\n", nc.DiscoveredServers()))
		}),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			cl.Logger.Error("client disconnected", slog.String("err", err.Error()))
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			cl.Logger.Info("client reconnected")
		}),
		nats.ClosedHandler(func(_ *nats.Conn) {
			cl.Logger.Info("client closed")
		}),
		nats.RetryOnFailedConnect(true),
		nats.PingInterval(2*time.Minute),
		nats.MaxPingsOutstanding(2),
		nats.ReconnectBufSize(8*1024*1024),
		nats.MaxReconnects(60),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		return err
	}
	js, err := jetstream.New(nc)
	if err != nil {
		return err
	}

	cl.JetStream = js
	cl.nc = nc

	return nil

}

func (cl *Client) Close() error {
	if cl.nc != nil {
		return cl.nc.Drain()
	}
	return nil
}

func (cl *Client) GetStream(ctx context.Context) (jetstream.Stream, error) {
	stream, err := cl.JetStream.Stream(ctx, StreamName)
	if errors.Is(err, jetstream.ErrStreamNotFound) {
		cl.Logger.Error(fmt.Sprintf("stream %s not found, creating new", StreamName))

		stream, err = cl.JetStream.CreateStream(ctx, cl.JetStreamCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create stream: %v", err)
		}

		cl.Logger.Info(fmt.Sprintf("stream %s created", StreamName))
		return stream, nil
	}

	cl.Logger.Info(fmt.Sprintf("stream %s found", StreamName))
	return stream, nil
}
