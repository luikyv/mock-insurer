package strutil

import (
	"math/rand"
)

const alphanumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func Random(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = alphanumeric[rand.Intn(len(alphanumeric))]
	}
	return string(result)
}
