package openlanv2

import (
	"time"
	"math/rand"
)

func GenUUID(n int) string {
	letters := []byte("0123456789abcdefghijklmnopqrstuvwxyz")
	buf := make([]byte, n)

	rand.Seed(time.Now().Unix())
    for i := range buf {
        buf[i] = letters[rand.Int63() % int64(len(letters))]
	}

	return string(buf)
}