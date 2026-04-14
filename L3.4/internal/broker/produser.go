package broker

import (
	"context"
	"encoding/json"
	"fmt"

	"imageProcessor/internal/dto"

	kafkav2 "github.com/wb-go/wbf/kafka/kafka-v2"
	"github.com/wb-go/wbf/logger"
)

// Producer описывает отправку сообщения в очередь.
type Producer interface {
	Send(ctx context.Context, msg dto.ProcessImageMessage) error
	Close() error
}

// KafkaProducer — обертка над wbf kafkav2 producer.
type KafkaProducer struct {
	producer *kafkav2.Producer
	log      logger.Logger
}

// NewKafkaProducer создает producer.
func NewKafkaProducer(brokers []string, topic string, log logger.Logger) *KafkaProducer {
	return &KafkaProducer{
		producer: kafkav2.NewProducer(brokers, topic, log),
		log:      log,
	}
}

// Send отправляет задачу на обработку изображения в Kafka.
func (p *KafkaProducer) Send(ctx context.Context, msg dto.ProcessImageMessage) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal kafka message: %w", err)
	}

	if err := p.producer.Send(ctx, []byte(msg.ID), payload); err != nil {
		return fmt.Errorf("send kafka message: %w", err)
	}

	p.log.LogAttrs(ctx, logger.InfoLevel, "image task sent to kafka",
		logger.String("image_id", msg.ID),
	)

	return nil
}

// Close закрывает producer.
func (p *KafkaProducer) Close() error {
	return p.producer.Close()
}
