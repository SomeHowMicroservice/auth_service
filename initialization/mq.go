package initialization

import (
	"fmt"
	"log"

	"github.com/SomeHowMicroservice/shm-be/auth/config"
	"github.com/rabbitmq/amqp091-go"
)

type MQConnection struct {
	Conn *amqp091.Connection
	Chann *amqp091.Channel
}

func InitMessageQueue(cfg *config.Config) (*MQConnection, error) {
	dsn := fmt.Sprintf("amqps://%s:%s@%s/%s",
		cfg.MessageQueue.RUser,
		cfg.MessageQueue.RPassword,
		cfg.MessageQueue.RHost,
		cfg.MessageQueue.RUser,
	)
	conn, err := amqp091.Dial(dsn)
	if err != nil {
		log.Fatalf("Kết nối RabbitMQ thất bại: %v", err)
	}
	chann, err := conn.Channel()
	if err != nil {
		log.Fatalf("Mở 1 kênh MQ thất bại: %v", err)
	}
	return &MQConnection{
		Conn: conn,
		Chann: chann,
	}, nil
}

func (mq *MQConnection) Close() {
	_ = mq.Chann.Close()
	_ = mq.Conn.Close()
}
