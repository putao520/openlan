package libol

import (
	"math/rand"
	"time"
)

func GenToken(n int) string {
	letters := []byte("0123456789abcdefghijklmnopqrstuvwxyz")
	buf := make([]byte, n)

	rand.Seed(time.Now().UnixNano())

	for i := range buf {
		buf[i] = letters[rand.Int63() % int64(len(letters))]
	}

	return string(buf)
}

func GenEthAddr(n int) []byte {
	if n == 0 {
		n = 6
	}

	data := make([]byte, n)
	rand.Seed(time.Now().UnixNano())
	for i := range data {
		data[i] = byte(rand.Uint32() & 0xFF)
	}

	data[0] &= 0xfe

	return data
}
