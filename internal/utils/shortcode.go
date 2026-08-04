package utils

import (
	"math/rand"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateShortCode(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	code := make([]byte, length)

	for i := range code {
		code[i] = charset[r.Intn(len(charset))]
	}

	return string(code)
}