package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	appName = "example-consumer"
)

func main() {
	kafkaTopic := os.Getenv("SOURCE_TOPIC")
	kafkaBrokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ";")

	// Kafka
	kafkaClient, err := NewKafkaConsumer(fmt.Sprintf("%s-group", appName), []string{kafkaTopic}, kafkaBrokers)
	if err != nil {
		log.Fatalf("failed to connect to Kafka: %s", err)
	}

	// Consume messages in a Go routine
	go func() {
		if err := Consume(context.Background(), kafkaClient); err != nil {
			log.Fatalf("failed to consume from Kafka: %s", err)
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT)
	go func() {
		<-done
		log.Println("finish application")
		os.Exit(0)
	}()

	for {
		// keep app running
	}
}

func NewKafkaConsumer(consumerGroup string, topics, brokers []string) (*kgo.Client, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(consumerGroup),
		kgo.ConsumeTopics(topics...),
		kgo.WithLogger(kgo.BasicLogger(os.Stderr, kgo.LogLevelInfo, nil)),
	}

	return kgo.NewClient(opts...)
}

// Consumer function example
func Consume(ctx context.Context, client *kgo.Client) error {
	for {
		fetches := client.PollFetches(ctx)
		if fetches.IsClientClosed() {
			return errors.New("kafka client is closed")
		}

		if errs := fetches.Errors(); len(errs) > 0 {
			return fmt.Errorf("multiple fetch errors found: %s", fmt.Sprint(errs))
		}

		// Call back function. It handles the record.
		fetches.EachRecord(func(r *kgo.Record) {
			log.Printf("Consumed message with key %s and value %v", string(r.Key), string(r.Value))
		})
	}
}
