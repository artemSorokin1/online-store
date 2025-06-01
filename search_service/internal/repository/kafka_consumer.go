package repository

import (
	"context"
	"encoding/json"
	"log"

	"searchservice/internal/models"

	"github.com/IBM/sarama"
)

type KafkaConsumer struct {
	broker  string
	topic   string
	groupID string
	handler func(*models.Product)
}

func NewKafkaConsumer(broker, topic, groupID string, handler func(*models.Product)) *KafkaConsumer {
	return &KafkaConsumer{
		broker:  broker,
		topic:   topic,
		groupID: groupID,
		handler: handler,
	}
}

func (kc *KafkaConsumer) Start(ctx context.Context) error {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumerGroup, err := sarama.NewConsumerGroup([]string{kc.broker}, kc.groupID, config)
	if err != nil {
		return err
	}
	defer consumerGroup.Close()

	loopHandler := &consumerGroupHandler{callback: kc.handler}

	for {
		if err := consumerGroup.Consume(ctx, []string{kc.topic}, loopHandler); err != nil {
			log.Printf("Error from Kafka consumer: %s", err)
		}
		if ctx.Err() != nil {
			return nil
		}
	}
}

type consumerGroupHandler struct {
	callback func(*models.Product)
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		var product models.Product
		if err := json.Unmarshal(message.Value, &product); err != nil {
			log.Printf("Error unmarshaling Kafka message: %s", err)
			session.MarkMessage(message, "")
			continue
		}
		h.callback(&product)
		session.MarkMessage(message, "")
	}
	return nil
}
