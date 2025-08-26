package smtp

type SMTPService interface {
	Send(to, subject, body string) error

	SendAuthEmail(to, subject, otp string) error
}
