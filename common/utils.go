package common

import (
	"math/rand"

	"golang.org/x/crypto/bcrypt"
)

func GenerateOTP(length int) string {
	const chars = "0123456789"
	otp := make([]byte, length)
	for i := range otp {
		otp[i] = chars[rand.Intn(len(chars))]
	}
	return string(otp)
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func VerifyPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
