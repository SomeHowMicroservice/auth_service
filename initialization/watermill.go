package initialization

import (
	"fmt"

	"github.com/SomeHowMicroservice/auth/common"
	"github.com/SomeHowMicroservice/auth/config"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
)

type WatermillConnection struct {
	Publisher  message.Publisher
	Subscriber message.Subscriber
}

func InitWatermill(cfg *config.Config, logger watermill.LoggerAdapter) (*WatermillConnection, error) {
	amqpConfig := amqp.NewDurablePubSubConfig(
		fmt.Sprintf("amqps://%s:%s@%s/%s",
			cfg.MessageQueue.RUser,
			cfg.MessageQueue.RPassword,
			cfg.MessageQueue.RHost,
			cfg.MessageQueue.RVhost,
		),
		nil,
	)

	amqpConfig.Exchange = amqp.ExchangeConfig{
		GenerateName: func(topic string) string {
			return common.Exchange
		},
		Type:    "topic",
		Durable: true,
	}

	amqpConfig.Publish.GenerateRoutingKey = func(topic string) string {
		return topic
	}

	amqpConfig.QueueBind = amqp.QueueBindConfig{
		GenerateRoutingKey: func(topic string) string {
			return topic
		},
	}

	amqpConfig.Queue = amqp.QueueConfig{
		GenerateName: func(topic string) string {
			return topic
		},
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
	}

	amqpConfig.Consume.Qos = amqp.QosConfig{
		PrefetchCount: 5,
	}

	amqpConfig.Publish.ConfirmDelivery = true

	publisher, err := amqp.NewPublisher(amqpConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("tạo publisher thất bại: %w", err)
	}

	subscriber, err := amqp.NewSubscriber(amqpConfig, logger)
	if err != nil {
		publisher.Close()
		return nil, fmt.Errorf("tạo subscriber thất bại: %w", err)
	}

	return &WatermillConnection{
		publisher,
		subscriber,
	}, nil
}

func (w *WatermillConnection) Close() {
	_ = w.Publisher.Close()
	_ = w.Subscriber.Close()
}