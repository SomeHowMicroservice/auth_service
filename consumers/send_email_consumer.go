package consumers

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/SomeHowMicroservice/shm-be/auth/smtp"
	"github.com/SomeHowMicroservice/shm-be/auth/common"
	"github.com/SomeHowMicroservice/shm-be/auth/initialization"
	"github.com/SomeHowMicroservice/shm-be/auth/mq"
)

func StartSendEmailConsumer(mqc *initialization.MQConnection, mailer smtp.SMTPService) {
	if err := mq.ConsumeMessage(mqc.Chann, common.QueueName, common.Exchange, common.RoutingKey, func(body []byte) error {
		var emailMsg common.AuthEmailMessage
		if err := json.Unmarshal(body, &emailMsg); err != nil {
			return fmt.Errorf("chuyển đổi tin nhắn email thất bại: %w", err)
		}
		
		if err := mailer.SendAuthEmail(emailMsg.To, emailMsg.Subject, emailMsg.Otp); err != nil {
			return fmt.Errorf("gửi email thất bại: %w", err)
		}

		log.Printf("Đã gửi email thành công tới: %s", emailMsg.To)
		return nil
	}); err != nil {
		log.Printf("Lỗi khởi tạo email consumer: %v", err)
	}
}