package subscriber

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

type Subscriber struct {
	common.JetStreamClient
}

func (s Subscriber) Start() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := s.JetStreamClient.JetStream.Stream(ctx, common.StreamName)
	if errors.Is(err, jetstream.ErrStreamNotFound) {
		fmt.Printf("stream %s not found, creating new\n", common.StreamName)
		stream, err = s.JetStreamClient.JetStream.CreateStream(ctx, s.JetStreamClient.JetStreamCfg)
		if err != nil {
			return fmt.Errorf("failed to create stream: %v", err)
		}
		fmt.Printf("stream %s created\n", common.StreamName)
	}

	fmt.Printf("stream %s found\n", common.StreamName)

	cons, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name:      "consumer-1",
		AckPolicy: jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return err
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)

	msgTicker := time.NewTicker(5 * time.Second)

outerLoop:
	for {
		select {
		case <-msgTicker.C:

			batch, err := cons.FetchNoWait(10)
			if err != nil {
				return err
			}
			for msg := range batch.Messages() {
				newMsg := common.Msg{}

				err = json.Unmarshal(msg.Data(), &newMsg)
				if err != nil {
					fmt.Printf("failed to unmarshal message: %v", err)
				} else {
					fmt.Printf("new message: %s\n", string(msg.Data()))
				}

				msg.Ack()
			}

			common.PrintStreamState(ctx, stream)
		case <-signalChan:
			break outerLoop
		}
	}

	fmt.Println("Finished")

	return nil
}
