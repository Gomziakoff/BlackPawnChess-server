package helpers

import (
	"math/rand"
	"os"
	"time"
)

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GenerateRandomString(n int) string {
	rand.NewSource(time.Now().UnixNano())

	// Define the character set
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Create a byte slice of length n
	b := make([]byte, n)

	// Fill the slice with random characters
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}
