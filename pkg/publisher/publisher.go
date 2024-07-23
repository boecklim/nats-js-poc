package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"nats-js-poc/pkg/common"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Publisher struct {
	common.JetStreamClient
}

func (p Publisher) Start() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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

			_, err = p.JetStreamClient.Js.Publish(ctx, common.Subject, msgBytes)
			if err != nil {
				return err
			}
			counter++
		case <-signalChan:
			break outerLoop
		}
	}

	fmt.Println("Finished")

	return nil
}
