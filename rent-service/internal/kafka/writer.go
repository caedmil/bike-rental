package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

// Writer интерфейс для Kafka Writer (для тестирования)
type Writer interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

// KafkaWriter обертка над kafka.Writer
type KafkaWriter struct {
	*kafka.Writer
}

func NewKafkaWriter(writer *kafka.Writer) Writer {
	return &KafkaWriter{Writer: writer}
}

func (w *KafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	return w.Writer.WriteMessages(ctx, msgs...)
}

func (w *KafkaWriter) Close() error {
	return w.Writer.Close()
}

