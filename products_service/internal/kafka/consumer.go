package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/Shopify/sarama"
)

type Consumer struct {
	consumer sarama.ConsumerGroup
	topic    string
	handler  EventHandler
}

type EventHandler interface {
	HandleProductEvent(ctx context.Context, event *ProductEvent) error
}

func NewConsumer(brokers []string, groupID, topic string, handler EventHandler) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания consumer: %v", err)
	}

	return &Consumer{
		consumer: consumer,
		topic:    topic,
		handler:  handler,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	topics := []string{c.topic}
	consumerGroup := &consumerGroupHandler{
		handler: c.handler,
	}

	for {
		err := c.consumer.Consume(ctx, topics, consumerGroup)
		if err != nil {
			return fmt.Errorf("ошибка потребления сообщений: %v", err)
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}

type consumerGroupHandler struct {
	handler EventHandler
}

func (h *consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		var event ProductEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			log.Printf("Ошибка десериализации сообщения: %v", err)
			continue
		}

		if err := h.handler.HandleProductEvent(session.Context(), &event); err != nil {
			log.Printf("Ошибка обработки события: %v", err)
			continue
		}

		session.MarkMessage(message, "")
	}
	return nil
} 