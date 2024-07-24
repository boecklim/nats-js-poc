package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"nats-js-poc/pkg/common"
	"time"

	"github.com/mroth/jitter"
)

type Subscriber struct {
	common.Client
}

func (s Subscriber) Start(ctx context.Context) error {

	waitForSeconds := 5
	s.Client.Logger.Info(fmt.Sprintf("wating for %d seconds", waitForSeconds))
	time.Sleep(time.Duration(waitForSeconds) * time.Second)

	err := s.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	stream, err := s.Client.GetStream(ctx)
	if err != nil {
		return fmt.Errorf("failed to get stream: %w", err)
	}

	cons, err := s.Client.GetConsumer(ctx, stream)
	if err != nil {
		return fmt.Errorf("failed to get consumer: %w", err)
	}

	if err != nil {
		return fmt.Errorf("failed to get consumer: %w", err)
	}

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

				// ... Do something with the message, then acknowledge

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
