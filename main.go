package main

import (
	"context"
	"errors"
	"github.com/nats-io/nats.go"
	"log"
	"log/slog"
	"nats-js-poc/pkg/common"
	"nats-js-poc/pkg/publisher"
	"nats-js-poc/pkg/subscriber"
	"os"
	"os/signal"
	"syscall"
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

	var client *common.Client
	var err error

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		signalChan := make(chan os.Signal, 1)
		//signal.Notify(signalChan, os.Interrupt) // Signal from Ctrl+C
		signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)
		<-signalChan

		logger.Info("Shutting down")
		cancel()
	}()

	arg := os.Args[1]
	switch arg {
	case "subscribe":

		client, err = common.NewJetStreamClient(url, logger)
		if err != nil {
			return err
		}
		p := subscriber.Subscriber{Client: *client}

		err = p.Start(ctx)
		if err != nil {
			return err
		}
		logger.Info("Finished subscribing")
	case "publish":

		client, err = common.NewJetStreamClient(url, logger)
		if err != nil {
			return err
		}
		s := publisher.Publisher{Client: *client}

		err = s.Start(ctx)
		if err != nil {
			return err
		}
		logger.Info("Finished publishing")
	default:
		return errors.New("either publish or subscribe")
	}

	return client.Close()
}
