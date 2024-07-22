package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func main() {
	err := run()
	if err != nil {
		log.Fatalf("failed to run: %v", err)
	}

	os.Exit(0)
}

func run() error {

	// Use the env variable if running in the container, otherwise use the default.
	url := "nats://127.0.0.1:4222"

	// Create an unauthenticated connection to NATS.
	nc, err := nats.Connect(url)
	if err != nil {
		return err
	}

	defer nc.Drain()

	// Access `JetStream` to use the JS APIs.
	js, err := jetstream.New(nc)
	if err != nil {
		return err
	}

	// ### Creating the stream
	// Define the stream configuration, specifying `WorkQueuePolicy` for
	// retention, and create the stream.
	cfg := jetstream.StreamConfig{
		Name:      "EVENTS",
		Retention: jetstream.WorkQueuePolicy,
		Subjects:  []string{"events.>"},
	}

	// JetStream API uses context for timeouts and cancellation.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := js.CreateStream(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create stream: %v", err)
	}
	fmt.Println("created the stream")

	// ### Queue messages
	// Publish a few messages.
	_, err = js.Publish(ctx, "events.us.page_loaded", nil)
	if err != nil {
		return err
	}
	_, err = js.Publish(ctx, "events.eu.mouse_clicked", nil)
	if err != nil {
		return err
	}
	_, err = js.Publish(ctx, "events.us.input_focused", nil)
	if err != nil {
		return err
	}
	fmt.Println("published 3 messages")

	// Checking the stream info, we see three messages have been queued.
	fmt.Println("# Stream info without any consumers")
	printStreamState(ctx, stream)

	// ### Adding a consumer
	// Now let's add a consumer and publish a few more messages.
	// [pull]: /examples/jetstream/pull-consumer/go
	cons, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name: "processor-1",
	})
	if err != nil {
		return err
	}

	// Fetch and ack the queued messages.
	msgs, err := cons.Fetch(3)
	if err != nil {
		return err
	}
	for msg := range msgs.Messages() {
		msg.DoubleAck(ctx)
	}

	// Checking the stream info again, we will notice no messages
	// are available.
	fmt.Println("\n# Stream info with one consumer")
	printStreamState(ctx, stream)

	// ### Exclusive non-filtered consumer
	// As noted in the description above, work-queue streams can only have
	// at most one consumer with interest on a subject at any given time.
	// Since the pull consumer above is not filtered, if we try to create
	// another one, it will fail.
	_, err = stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name: "processor-2",
	})
	if err != nil {
		return err
	}
	fmt.Println("\n# Create an overlapping consumer")
	fmt.Println(err)

	// However if we delete the first one, we can then add the new one.
	stream.DeleteConsumer(ctx, "processor-1")

	_, err = stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name: "processor-2",
	})
	if err != nil {
		return err
	}
	fmt.Printf("created the new consumer? %v\n", err == nil)
	stream.DeleteConsumer(ctx, "processor-2")

	// ### Multiple filtered consumers
	// To create multiple consumers, a subject filter needs to be applied.
	// For this example, we could scope each consumer to the geo that the
	// event was published from, in this case `us` or `eu`.
	fmt.Println("\n# Create non-overlapping consumers")
	cons1, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name:          "processor-us",
		FilterSubject: "events.us.>",
	})
	if err != nil {
		return err
	}
	cons2, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name:          "processor-eu",
		FilterSubject: "events.eu.>",
	})
	if err != nil {
		return err
	}
	_, err = js.Publish(ctx, "events.eu.mouse_clicked", nil)
	if err != nil {
		return err
	}
	_, err = js.Publish(ctx, "events.us.page_loaded", nil)
	if err != nil {
		return err
	}
	_, err = js.Publish(ctx, "events.us.input_focused", nil)
	if err != nil {
		return err
	}
	_, err = js.Publish(ctx, "events.eu.page_loaded", nil)
	if err != nil {
		return err
	}
	fmt.Println("published 4 messages")

	msgs, err = cons1.Fetch(2)
	if err != nil {
		return err
	}

	for msg := range msgs.Messages() {
		fmt.Printf("us sub got: %s\n", msg.Subject())
		msg.Ack()
	}

	msgs, err = cons2.Fetch(2)
	if err != nil {
		return err
	}

	for msg := range msgs.Messages() {
		fmt.Printf("eu sub got: %s\n", msg.Subject())
		msg.Ack()
	}

	return nil
}

// This is just a helper function to print the stream state info 😉.
func printStreamState(ctx context.Context, stream jetstream.Stream) {
	info, _ := stream.Info(ctx)
	b, _ := json.MarshalIndent(info.State, "", " ")
	fmt.Println(string(b))
}
