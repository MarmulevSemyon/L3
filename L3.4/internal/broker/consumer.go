package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"imageProcessor/internal/dto"

	"github.com/segmentio/kafka-go"
	"github.com/wb-go/wbf/kafka/dlq"
	kafkav2 "github.com/wb-go/wbf/kafka/kafka-v2"
	"github.com/wb-go/wbf/logger"
)

// MessageProcessor описывает любой объект,
// который умеет обрабатывать сообщение из Kafka.
type MessageProcessor interface {
	ProcessMessage(ctx context.Context, msg dto.ProcessImageMessage) error
}

// Consumer читает сообщения из Kafka и передает их обработчику.
type Consumer struct {
	consumer  *kafkav2.Consumer
	processor *kafkav2.Processor
	log       logger.Logger
}

// NewConsumer создает consumer и processor.
func NewConsumer(
	brokers []string,
	topic string,
	groupID string,
	dlqTopic string,
	log logger.Logger,
	processorHandler MessageProcessor,
) (*Consumer, error) {
	mainConsumer := kafkav2.NewConsumer(brokers, topic, groupID, log)

	var dlqClient *dlq.DLQ
	if dlqTopic != "" {
		dlqProducer := kafkav2.NewProducer(brokers, dlqTopic, log)
		dlqClient = dlq.New(dlqProducer, log)
	}

	processor, err := kafkav2.NewProcessor(
		mainConsumer,
		dlqClient,
		log,
		kafkav2.MaxAttempts(3),
		kafkav2.BaseRetryDelay(200*time.Millisecond),
		kafkav2.MaxRetryDelay(2*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("create kafka processor: %w", err)
	}

	c := &Consumer{
		consumer:  mainConsumer,
		processor: processor,
		log:       log,
	}

	c.processor.Start(context.Background(), c.makeHandler(processorHandler))

	return c, nil
}

func (c *Consumer) makeHandler(processorHandler MessageProcessor) func(ctx context.Context, msg kafka.Message) error {
	return func(ctx context.Context, msg kafka.Message) error {
		var payload dto.ProcessImageMessage
		if err := json.Unmarshal(msg.Value, &payload); err != nil {
			return fmt.Errorf("unmarshal kafka message: %w", err)
		}

		c.log.LogAttrs(ctx, logger.InfoLevel, "kafka message received",
			logger.String("image_id", payload.ID),
			logger.String("topic", msg.Topic),
		)

		if err := processorHandler.ProcessMessage(ctx, payload); err != nil {
			return fmt.Errorf("process image message: %w", err)
		}

		return nil
	}
}

// Close закрывает consumer.
func (c *Consumer) Close() error {
	return c.consumer.Close()
}
