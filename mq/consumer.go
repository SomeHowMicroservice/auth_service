package mq

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/SomeHowMicroservice/auth/common"
	"github.com/SomeHowMicroservice/auth/smtp"
	"github.com/ThreeDotsLabs/watermill/message"
)

func RegisterSendEmailConsumer(router *message.Router, subscriber message.Subscriber, mailer smtp.SMTPService) {
	router.AddConsumerHandler(
		"send_email_handler",
		common.SendTopic,
		subscriber,
		message.NoPublishHandlerFunc(func(msg *message.Message) error {
			return handleSendEmail(msg, mailer)
		}),
	)
}

func handleSendEmail(msg *message.Message, mailer smtp.SMTPService) error {
	var emailMsg common.AuthEmailMessage
	if err := json.Unmarshal(msg.Payload, &emailMsg); err != nil {
		return fmt.Errorf("chuyển đổi tin nhắn email thất bại: %w", err)
	}

	if err := mailer.SendAuthEmail(emailMsg.To, emailMsg.Subject, emailMsg.Otp); err != nil {
		return fmt.Errorf("gửi email thất bại: %w", err)
	}
	log.Printf("Đã gửi email thành công tới: %s", emailMsg.To)
	
	return nil
}
