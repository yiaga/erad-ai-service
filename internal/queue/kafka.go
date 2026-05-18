package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
)

type KafkaQueue struct {
	writer *kafka.Writer
	reader *kafka.Reader
	topic  string
}

func NewKafkaQueue(brokers []string, topic string, groupID string) *KafkaQueue {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &KafkaQueue{
		writer: writer,
		reader: reader,
		topic:  topic,
	}
}

func (q *KafkaQueue) Publish(ctx context.Context, payload JobPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return q.writer.WriteMessages(ctx, kafka.Message{
		Value: data,
	})
}

func (q *KafkaQueue) Subscribe(ctx context.Context, handler func(ctx context.Context, payload JobPayload) error) error {
	for {
		m, err := q.reader.ReadMessage(ctx)
		if err != nil {
			return err
		}

		var payload JobPayload
		if err := json.Unmarshal(m.Value, &payload); err != nil {
			fmt.Printf("Error unmarshaling message: %v\n", err)
			continue
		}

		if err := handler(ctx, payload); err != nil {
			fmt.Printf("Error handling message: %v\n", err)
			// In a real system, we might want to retry or send to a DLQ
		}
	}
}

func (q *KafkaQueue) Close() error {
	if err := q.writer.Close(); err != nil {
		return err
	}
	return q.reader.Close()
}
