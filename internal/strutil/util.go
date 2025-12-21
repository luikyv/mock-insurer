package strutil

import (
	"crypto/rand"
	"encoding/base64"
)

func Random(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(bytes)
}
