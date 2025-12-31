package strutil

import (
	"crypto/rand"
	"math/big"
)

const alphanumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func Random(length int) string {
	result := make([]byte, length)
	max := big.NewInt(int64(len(alphanumeric)))
	for i := range result {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			panic(err)
		}
		result[i] = alphanumeric[n.Int64()]
	}
	return string(result)
}
