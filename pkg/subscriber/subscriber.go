package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"nats-js-poc/pkg/common"
	"time"

	"github.com/mroth/jitter"
	"github.com/nats-io/nats.go/jetstream"
)

type Subscriber struct {
	common.Client
}

const consumerName = "consumer-1"

func (s Subscriber) Start(ctx context.Context) error {

	waitForSeconds := 10
	s.Client.Logger.Info(fmt.Sprintf("wating for %d seconds", waitForSeconds))
	time.Sleep(time.Duration(waitForSeconds) * time.Second)

	err := s.Connect()
	if err != nil {
		return err
	}

	stream, err := s.Client.GetStream(ctx)
	if err != nil {
		return err
	}
	s.Client.Logger.Info("got stream")

	cons, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name:      consumerName,
		AckPolicy: jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return err
	}

	s.Client.Logger.Info("got consumer")

	s.Client.Logger.Info("starting consuming")

	msgTicker := jitter.NewTicker(3*time.Second, 0.5)
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
					s.Client.Logger.Info("message received", slog.String("msg", newMsg.Msg), slog.String("uuid", newMsg.UUID))
				}

				err = msg.Ack()
				if err != nil {
					s.Client.Logger.Error("failed to ack message", slog.String("err", err.Error()))
				}
			}

		case <-ctx.Done():
			break outerLoop
		}
	}

	s.Client.Logger.Info("Finished")

	return nil
}
