package main

import (
	"errors"
	"log"
	"log/slog"
	"nats-js-poc/pkg/common"
	"nats-js-poc/pkg/publisher"
	"nats-js-poc/pkg/subscriber"
	"os"

	"github.com/nats-io/nats.go"
)

func main() {
	err := run()
	if err != nil {
		log.Fatalf("failed to run: %v", err)
	}

	os.Exit(0)
}

type Msg struct {
	Msg string `json:"name"`
	ID  int    `json:"id"`
}

func run() error {

	// Use the env variable if running in the container, otherwise use the default.
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = nats.DefaultURL
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	client, err := common.NewJetStreamClient(url, logger)
	if err != nil {
		return err
	}

	arg := os.Args[1]
	switch arg {
	case "subscribe":

		p := subscriber.Subscriber{Client: *client}

		err = p.Start()
		if err != nil {
			return err
		}
		logger.Info("Finished subscribing")
	case "publish":

		s := publisher.Publisher{Client: *client}

		err = s.Start()
		if err != nil {
			return err
		}
		logger.Info("Finished publishing")
	default:
		return errors.New("either publish or subscribe")
	}

	return client.Close()
}
