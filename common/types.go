package common

type RegistrationData struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	Otp      string `json:"otp"`
	Attempts int    `json:"attempts"`
}

type AuthEmailMessage struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Otp     string `json:"otp"`
}

type ForgotPasswordData struct {
	Email    string `json:"email"`
	Otp      string `json:"otp"`
	Attempts int    `json:"attempts"`
}
