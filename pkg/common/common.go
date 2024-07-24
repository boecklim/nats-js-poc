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

const (
	streamDesc   = "Stream for testing JetStream as message queue"
	streamName   = "test-stream-1"
	Subject      = "test-subject"
	consumerName = "consumer-1"
)

type Msg struct {
	Msg  string `json:"name"`
	UUID string `json:"uuid"`
}

type Client struct {
	Js     jetstream.JetStream
	jsCfg  jetstream.StreamConfig
	nc     *nats.Conn
	Logger *slog.Logger
	url    string
}

func NewJetStreamClient(url string, logger *slog.Logger) (*Client, error) {

	cfg := jetstream.StreamConfig{
		Name:        streamName,
		Description: streamDesc,
		Subjects:    []string{Subject},
		Retention:   jetstream.WorkQueuePolicy,
		MaxBytes:    50000, // 5 GB
		Discard:     jetstream.DiscardOld,
		MaxAge:      10 * time.Minute,
		Storage:     jetstream.MemoryStorage,
		NoAck:       false,
	}

	p := &Client{
		jsCfg:  cfg,
		Logger: logger,
		url:    url,
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
			cl.Logger.Info(fmt.Sprintf("======= Known servers: %v\n", nc.Servers()))
			cl.Logger.Info(fmt.Sprintf("======= Discovered servers: %v\n", nc.DiscoveredServers()))
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

	cl.Js = js
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
	streamCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	stream, err := cl.Js.Stream(streamCtx, streamName)
	if errors.Is(err, jetstream.ErrStreamNotFound) {
		cl.Logger.Error(fmt.Sprintf("stream %s not found, creating new", streamName))

		stream, err = cl.Js.CreateStream(streamCtx, cl.jsCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create stream: %v", err)
		}

		cl.Logger.Info(fmt.Sprintf("stream %s created", streamName))
		return stream, nil
	} else if err != nil {
		return nil, err
	}

	cl.Logger.Info(fmt.Sprintf("stream %s found", streamName))
	return stream, nil
}

func (cl *Client) GetConsumer(ctx context.Context, stream jetstream.Stream) (jetstream.Consumer, error) {

	consCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cons, err := stream.Consumer(consCtx, consumerName)
	if errors.Is(err, jetstream.ErrConsumerNotFound) {
		cl.Logger.Error(fmt.Sprintf("consumer %s not found, creating new", consumerName))
		cons, err = stream.CreateConsumer(consCtx, jetstream.ConsumerConfig{
			Durable:   consumerName,
			AckPolicy: jetstream.AckExplicitPolicy,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create consumer: %v", err)
		}

		cl.Logger.Info(fmt.Sprintf("consumer %s created", consumerName))
		return cons, nil

	} else if err != nil {
		return nil, err
	}

	cl.Logger.Info(fmt.Sprintf("consumer %s found", consumerName))
	return cons, err
}
