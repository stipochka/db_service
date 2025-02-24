package consumer

import (
	"context"
	"log/slog"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/db_service/internal/service"
)

type KafkaMessageConsumer struct {
	kafkaService *service.RecordSaverService
	consumer     *kafka.Consumer
	log          *slog.Logger
}

func NewKafkaMessageConsumer(address, groupID, topic string, kafkaService *service.RecordSaverService, log slog.Logger) (*KafkaMessageConsumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": address,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, err
	}

	err = c.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		return nil, err
	}

	return &KafkaMessageConsumer{
		kafkaService: kafkaService,
		consumer:     c,
		log:          &log,
	}, nil
}

func (k *KafkaMessageConsumer) Consume(ctx context.Context) error {
	const op = "consumer.Consume"

	defer k.consumer.Close()

	log := k.log.With(slog.String("op", op))

	for {
		select {
		case <-ctx.Done():
			log.Info("stopping consumer")
			return nil
		default:
			msg, err := k.consumer.ReadMessage(-1)
			if err != nil {
				log.Error("failed to get message", slog.String("error", err.Error()))
				continue
			}

			slog.Info("message received", slog.String("key", string(msg.Key)), slog.String("value", string(msg.Value)))
			if string(msg.Value) == "" {
				continue
			}

			id, err := k.kafkaService.CreateRecord(context.Background(), msg.Key, msg.Value)
			if err != nil {
				log.Error("failed to create record", slog.String("error", err.Error()))
				continue
			}

			log.Info("created record with id", slog.Int("id", id))
		}

	}
}
