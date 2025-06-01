package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Shopify/sarama"
)

type Producer struct {
	producer sarama.SyncProducer
	topic    string
}

type ProductEvent struct {
	EventType string `json:"event_type"` // "created", "updated", "deleted"
	ProductID string `json:"product_id"`
	Name      string `json:"name"`
	Description string `json:"description"`
	Tags      []string `json:"tags"`
	Seller    string `json:"seller"`
}

func NewProducer(brokers []string, topic string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания producer: %v", err)
	}

	return &Producer{
		producer: producer,
		topic:    topic,
	}, nil
}

func (p *Producer) SendProductEvent(ctx context.Context, event *ProductEvent) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("ошибка сериализации события: %v", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.StringEncoder(eventJSON),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("ошибка отправки сообщения: %v", err)
	}

	log.Printf("Сообщение отправлено в partition %d с offset %d", partition, offset)
	return nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
} 