package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"nats-js-poc/pkg/common"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mroth/jitter"
	"github.com/nats-io/nats.go/jetstream"
)

type Subscriber struct {
	common.Client
}

func (s Subscriber) Start() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := s.Client.GetStream(ctx)
	if err != nil {
		return err
	}
	s.Client.Logger.Info("got stream")

	cons, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name:      "consumer-1",
		AckPolicy: jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return err
	}

	s.Client.Logger.Info("got consumer")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)

	msgTicker := jitter.NewTicker(3*time.Second, 0.5)

	s.Client.Logger.Info("starting consuming")

outerLoop:
	for {
		select {
		case <-msgTicker.C:

			batch, err := cons.Fetch(1)
			if err != nil {
				return err
			}
			for msg := range batch.Messages() {
				newMsg := common.Msg{}

				err = json.Unmarshal(msg.Data(), &newMsg)
				if err != nil {
					fmt.Printf("failed to unmarshal message: %v", err)
					s.Client.Logger.Error("failed to unmarshal message", slog.String("err", err.Error()))
				} else {
					s.Client.Logger.Info("new msg", slog.String("msg", string(msg.Data())))
				}

				msg.Ack()
			}

		case <-signalChan:
			break outerLoop
		}
	}

	s.Client.Logger.Info("Finished")

	return nil
}
