package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"

	"auth-service/pkg/domain"
)

type Producer struct {
	emailWriter *kafka.Writer
	userWriter  *kafka.Writer
}

func NewProducer(brokers []string, emailTopic, userTopic string) *Producer {
	return &Producer{
		emailWriter: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    emailTopic,
			Balancer: &kafka.LeastBytes{},
		},
		userWriter: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    userTopic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *Producer) SendEmail(ctx context.Context, email domain.EmailMessage) error {
	value, err := json.Marshal(email)
	if err != nil {
		return err
	}

	return p.emailWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(email.ID),
		Value: value,
	})
}

func (p *Producer) SendUserRegistered(ctx context.Context, event domain.UserRegisteredEvent) error {
	value, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.userWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.UserID),
		Value: value,
	})
}

func (p *Producer) Close() error {
	if err := p.emailWriter.Close(); err != nil {
		return err
	}
	return p.userWriter.Close()
}
