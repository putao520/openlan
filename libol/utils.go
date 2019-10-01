package libol

import (
	"bytes"
	"encoding/json"
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

func Marshal(v interface {}, pretty bool) (string, error) {
	str , err := json.Marshal(v)
	if err != nil {
		Error("Marshal error: %s" , err)
		return "", err
	}

	if !pretty {
		return string(str), nil
	}

	var out bytes.Buffer

	if err := json.Indent(&out, str, "", "  "); err != nil {
		return string(str), nil
	}

	return out.String(), nil
}