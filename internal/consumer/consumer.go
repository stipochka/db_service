package consumer

import (
	"context"
	"log/slog"

	"github.com/db_service/internal/service"
)

type Consumer interface {
	Consume(ctx context.Context) error
}

type KafkaConsumer struct {
	Consumer
}

func NewKafkaConsumer(address string, groupID string, topic string, kafkaService *service.RecordSaverService, log slog.Logger) (*KafkaConsumer, error) {
	consumer, err := NewKafkaMessageConsumer(address, groupID, topic, kafkaService, log)
	if err != nil {
		return nil, err
	}

	return &KafkaConsumer{
		Consumer: consumer,
	}, nil
}
