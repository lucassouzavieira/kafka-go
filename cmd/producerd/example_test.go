package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/twmb/franz-go/pkg/kfake"
	"github.com/twmb/franz-go/pkg/kgo"
)

var brokerAddr string

func TestMain(m *testing.M) {
	// Setup test environment
	// Create an mocked Kafka cluster
	cluster, err := kfake.NewCluster(
		kfake.Ports(9092),
		kfake.NumBrokers(1),
		kfake.AllowAutoTopicCreation(),
	)
	if err != nil {
		panic(err)
	}
	defer cluster.Close()
	for _, address := range cluster.ListenAddrs() {
		brokerAddr = address
	}

	code := m.Run()
	os.Exit(code)
}

func TestShouldDoSomething(t *testing.T) {
	// Connect to mock Kafka cluster
	client, err := NewTestKafkaConsumer("test.group", []string{"test.topic"}, []string{brokerAddr})
	assert.NoError(t, err)

	// Create test messages
	produceTestMessage("test.topic", []byte("Hello"))
	produceTestMessage("test.topic", []byte("World"))

	// Need to consume the test messages - Call your consumer func
	err = Consume(context.Background(), client)
	assert.NoError(t, err)
}

func produceTestMessage(topic string, message []byte) error {
	opts := []kgo.Opt{
		kgo.SeedBrokers([]string{brokerAddr}...),
		kgo.AllowAutoTopicCreation(),
		kgo.WithLogger(kgo.BasicLogger(os.Stderr, kgo.LogLevelInfo, nil)),
	}
	producer, err := kgo.NewClient(opts...)
	if err != nil {
		return err
	}

	record := kgo.Record{
		Topic: topic,
		Key:   []byte(uuid.New().String()),
		Value: message,
	}

	producer.Produce(context.Background(), &record, func(r *kgo.Record, err error) {
		if err != nil {
			fmt.Print("failed to produce record")
		}
	})
	return nil
}

func NewTestKafkaConsumer(consumerGroup string, topics, brokers []string) (*kgo.Client, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(consumerGroup),
		kgo.ConsumeTopics(topics...),
		kgo.WithLogger(kgo.BasicLogger(os.Stderr, kgo.LogLevelInfo, nil)),
	}

	return kgo.NewClient(opts...)
}

func Consume(ctx context.Context, client *kgo.Client) error {
	fetches := client.PollFetches(ctx)
	if fetches.IsClientClosed() {
		return errors.New("kafka client is closed")
	}

	if errs := fetches.Errors(); len(errs) > 0 {
		return fmt.Errorf("multiple fetch errors found: %s", fmt.Sprint(errs))
	}

	// Call back function. It handles the record.
	fetches.EachRecord(func(r *kgo.Record) {
		// Action
		log.Printf("Consumed message with key %s and value %v", string(r.Key), string(r.Value))
	})

	return nil
}
