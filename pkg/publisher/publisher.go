package publisher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"nats-js-poc/pkg/common"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

type Publisher struct {
	common.JetStreamClient
}

func (p Publisher) Start() error {
	ctx := context.Background()

	stream, err := p.JetStreamClient.JetStream.Stream(ctx, common.StreamName)
	if errors.Is(err, jetstream.ErrStreamNotFound) {
		fmt.Printf("stream %s not found, creating new\n", common.StreamName)
		stream, err = p.JetStreamClient.JetStream.CreateStream(ctx, p.JetStreamClient.JetStreamCfg)
		if err != nil {
			return fmt.Errorf("failed to create stream: %v", err)
		}
		fmt.Printf("stream %s created\n", common.StreamName)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)

	msgTicker := time.NewTicker(1 * time.Second)
	counter := 0
outerLoop:
	for {
		select {
		case <-msgTicker.C:
			msgBytes, err := json.Marshal(common.Msg{Msg: time.Now().Format(time.RFC3339), ID: counter})
			if err != nil {
				return err
			}

			_, err = p.JetStreamClient.JetStream.Publish(ctx, common.Subject, msgBytes)
			if err != nil {
				return err
			}
			counter++

			common.PrintStreamState(ctx, stream)
		case <-signalChan:
			break outerLoop
		}
	}

	fmt.Println("Finished")

	return nil
}
