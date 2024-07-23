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

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Publisher struct {
	js  jetstream.JetStream
	cfg jetstream.StreamConfig
}
type Msg struct {
	Msg string `json:"name"`
	ID  int    `json:"id"`
}

func NewPublisher(url string) (*Publisher, error) {

	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	defer nc.Drain()

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}

	cfg := jetstream.StreamConfig{
		Name:      common.StreamName,
		Retention: jetstream.WorkQueuePolicy,
		Subjects:  []string{},
		Storage:   jetstream.MemoryStorage,
	}

	p := &Publisher{
		js:  js,
		cfg: cfg,
	}

	return p, nil
}

func (p Publisher) Start() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := p.js.Stream(ctx, common.StreamName)
	if errors.Is(err, jetstream.ErrStreamNotFound) {
		fmt.Printf("stream %s not found, creating new\n", common.StreamName)
		stream, err = p.js.CreateStream(ctx, p.cfg)
		if err != nil {
			return fmt.Errorf("failed to create stream: %v", err)
		}
		fmt.Printf("stream %s created\n", common.StreamName)
	}

	fmt.Printf("stream %s found\n", common.StreamName)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)

	msgTicker := time.NewTicker(1 * time.Second)
	counter := 0
outerLoop:
	for {
		select {
		case <-msgTicker.C:
			msgBytes, err := json.Marshal(Msg{Msg: time.Now().Format(time.RFC3339), ID: counter})
			if err != nil {
				return err
			}

			_, err = p.js.Publish(ctx, common.Subject, msgBytes)
			if err != nil {
				return err
			}
			counter++
			printStreamState(ctx, stream)
		case <-signalChan:
			break outerLoop
		}
	}

	fmt.Println("Finished")

	return nil
}
func printStreamState(ctx context.Context, stream jetstream.Stream) error {
	info, err := stream.Info(ctx)
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(info.State, "", " ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))

	return nil
}
