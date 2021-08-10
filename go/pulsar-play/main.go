package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

func main() {
	ctx := context.Background()

	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               "pulsar://localhost:6650",
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
	})
	exitOnError(err)
	defer client.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		consumer, err := client.Subscribe(pulsar.ConsumerOptions{
			Topic:            "persistent://manning/chapter03/example-topic",
			SubscriptionName: "my0subscription",
		})
		exitOnError(err)
		defer consumer.Close()

		for i := 0; i < 10; i++ {
			msg, err := consumer.Receive(ctx)
			exitOnError(err)

			fmt.Printf("Received message: msgId: %#v -- content: '%s'\n", msg.ID(), string(msg.Payload()))

			consumer.Ack(msg)
		}

		exitOnError(consumer.Unsubscribe())
		wg.Done()
	}()

	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: "persistent://manning/chapter03/example-topic",
	})
	exitOnError(err)
	defer producer.Close()

	for i := 0; i < 10; i++ {
		msg := &pulsar.ProducerMessage{
			Key:     "some-key",
			Payload: []byte(fmt.Sprintf("Message # %d", i)),
			Properties: map[string]string{
				"property_1": "value_1",
				"property_2": "value_2",
			},
		}
		_, err = producer.Send(ctx, msg)
		exitOnError(err)
		time.Sleep(500 * time.Millisecond)
	}

	wg.Wait()
}

func exitOnError(err error) {
	if err != nil {
		log.Fatalf("Fatal error encountered: %v", err)
	}
}
