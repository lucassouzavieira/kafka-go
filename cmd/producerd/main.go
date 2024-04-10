package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	appName = "example-producer"
)

func main() {
	kafkaTopic := os.Getenv("SINK_TOPIC")
	kafkaBrokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ";")

	// Kafka
	kafkaClient, err := NewKafkaProducer(kafkaBrokers)
	if err != nil {
		log.Fatalf("failed to connect to Kafka: %s", err)
	}
	defer kafkaClient.Close()

	// Produce messages in a Go routine
	go func() {
		for {
			message := []byte(fmt.Sprintf("new message created at %s", time.Now().Format(time.RFC3339)))

			record := kgo.Record{
				Key:   []byte(uuid.New().String()),
				Value: []byte(message),
				Headers: []kgo.RecordHeader{
					{
						Key:   "application",
						Value: []byte(appName),
					},
				},
				Timestamp: time.Now(),
				Topic:     kafkaTopic,
			}

			kafkaClient.Produce(context.Background(), &record, func(r *kgo.Record, err error) {
				if err != nil {
					log.Printf("failed to produce message with key %s: %s", string(r.Key), err.Error())
				}
			})

			time.Sleep(300 * time.Millisecond)
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

func NewKafkaProducer(brokers []string) (*kgo.Client, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
		kgo.AllowAutoTopicCreation(),
		kgo.WithLogger(kgo.BasicLogger(os.Stderr, kgo.LogLevelInfo, nil)),
	}

	return kgo.NewClient(opts...)
}
