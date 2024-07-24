package publisher

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
)

type Publisher struct {
	common.Client
}

func (p Publisher) Start() error {
	ctx := context.Background()

	stream, err := p.Client.GetStream(ctx)
	if err != nil {
		return err
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)

	msgTicker := jitter.NewTicker(3*time.Second, 0.5)
	counter := 0
outerLoop:
	for {
		select {
		case <-msgTicker.C:
			msgBytes, err := json.Marshal(common.Msg{Msg: time.Now().Format(time.RFC3339), ID: counter})
			if err != nil {
				return err
			}

			_, err = p.Client.JetStream.Publish(ctx, common.Subject, msgBytes)
			if err != nil {
				return err
			}
			p.Client.Logger.Info("published message")
			counter++

			info, err := stream.Info(ctx)
			if err != nil {
				p.Client.Logger.Error("failed to get info", slog.String("err", err.Error()))
				continue
			}

			p.Client.Logger.Info("info",
				slog.Int("cons", info.State.Consumers),
				slog.Int("subjects", int(info.State.NumSubjects)),
				slog.Int("msgs", int(info.State.Msgs)),
			)

		case <-signalChan:
			break outerLoop
		}
	}

	fmt.Println("Finished")

	return nil
}
