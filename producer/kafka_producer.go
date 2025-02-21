package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kameshsampath/balloon-popper-server/models"
	"github.com/twmb/franz-go/pkg/kgo"
)

type KafkaScoreProducer struct {
	client *kgo.Client
	topic  string
}

func NewKafkaScoreProducer(bootstrapServers, topic string) (*KafkaScoreProducer, error) {
	// Create Kafka client configuration
	opts := []kgo.Opt{
		kgo.SeedBrokers(bootstrapServers),
		kgo.ProducerLinger(time.Millisecond * 100), // Wait up to 100ms to batch records
		//this is fail if quorum has just one broker
		//kgo.RequiredAcks(kgo.LeaderAck()),          // Wait for leader acknowledgment
	}

	// Create the Kafka client
	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka client: %w", err)
	}

	return &KafkaScoreProducer{
		client: client,
		topic:  topic,
	}, nil
}

func (k *KafkaScoreProducer) Start() error {
	// Ping the brokers to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := k.client.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping kafka brokers: %w", err)
	}
	return nil
}

func (k *KafkaScoreProducer) Stop() error {
	if k.client != nil {
		k.client.Close()
	}
	return nil
}

func (k *KafkaScoreProducer) SendScore(ctx context.Context, event *models.GameEvent) error {
	if k.client == nil {
		return fmt.Errorf("kafka client not initialized")
	}

	// Convert event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create record
	record := &kgo.Record{
		Topic: k.topic,
		Value: eventJSON,
		Key:   []byte(event.Player), // Using player name as key for partitioning
	}

	// Produce the record
	result := k.client.ProduceSync(ctx, record)
	if result.FirstErr() != nil {
		return fmt.Errorf("failed to produce message: %w", result.FirstErr())
	}

	return nil
}

// SendScoreBatch sends multiple game events in a batch
func (k *KafkaScoreProducer) SendScoreBatch(ctx context.Context, events []*models.GameEvent) error {
	if k.client == nil {
		return fmt.Errorf("kafka client not initialized")
	}

	records := make([]*kgo.Record, len(events))
	for i, event := range events {
		eventJSON, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}

		records[i] = &kgo.Record{
			Topic: k.topic,
			Value: eventJSON,
			Key:   []byte(event.Player),
		}
	}

	// Produce all records
	results := k.client.ProduceSync(ctx, records...)
	for _, result := range results {
		if result.Err != nil {
			return fmt.Errorf("failed to produce message: %w", result.Err)
		}
	}

	return nil
}
