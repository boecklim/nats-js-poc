package main

import (
	"errors"
	"fmt"
	"log"
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

	client, err := common.NewJetStreamClient(url)
	if err != nil {
		return err
	}

	arg := os.Args[1]
	switch arg {
	case "subscribe":

		p := subscriber.Subscriber{JetStreamClient: *client}

		err = p.Start()
		if err != nil {
			return err
		}
		fmt.Println("Finished subscribing")
	case "publish":

		s := publisher.Publisher{JetStreamClient: *client}

		err = s.Start()
		if err != nil {
			return err
		}
		fmt.Println("Finished publishing")
	default:
		return errors.New("either publish or subscribe")
	}

	return client.Close()
}
