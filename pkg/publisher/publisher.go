package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"nats-js-poc/pkg/common"
	"time"

	"github.com/google/uuid"
	"github.com/mroth/jitter"
)

type Publisher struct {
	common.Client
}

func (p Publisher) Start(ctx context.Context) error {
	err := p.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	stream, err := p.Client.GetStream(ctx)
	if err != nil {
		return fmt.Errorf("failed to get stream: %w", err)
	}

	p.Client.Logger.Info("starting publishing")

	msgTicker := jitter.NewTicker(3*time.Second, 0.5)
outerLoop:
	for {
		select {
		case <-msgTicker.C:
			msgBytes, err := json.Marshal(common.Msg{Msg: time.Now().Format(time.RFC3339), UUID: uuid.New().String()})
			if err != nil {
				return err
			}

			_, err = p.Client.JetStream.Publish(ctx, common.Subject, msgBytes)
			if err != nil {
				return err
			}

			info, err := stream.Info(ctx)
			if err != nil {
				p.Client.Logger.Error("failed to get info", slog.String("err", err.Error()))
				continue
			}

			p.Client.Logger.Info("message published",
				slog.Int("cons", info.State.Consumers),
				slog.Int("subjects", int(info.State.NumSubjects)),
				slog.Int("msgs", int(info.State.Msgs)),
			)

		case <-ctx.Done():
			break outerLoop
		}
	}

	fmt.Println("Finished")

	return nil
}
