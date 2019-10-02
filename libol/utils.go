package libol

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"os"
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

func MarshalSave(v interface {}, file string, pretty bool) error {
	f, err := os.OpenFile(file, os.O_RDWR | os.O_TRUNC | os.O_CREATE, 0600)
	defer f.Close()
	if err != nil {
		Error("MarshalSave: %s", err)
		return err
	}

	str, err := Marshal(v, true)
	if err != nil {
		Error("MarshalSave error: %s" , err)
		return err
	}

	if _, err := f.Write([]byte(str)); err != nil {
		Error("MarshalSave: %s", err)
		return err
	}

	return nil
}

func UnmarshalLoad(v interface {}, file string) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return Errer("UnmarshalLoad: file:%s does not exist", file)
	}

	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return Errer("UnmarshalLoad: file:%s %s", file, err)
	}

	if err := json.Unmarshal([]byte(contents), v); err != nil {
		return Errer("UnmarshalLoad: %s", err)
	}

	return nil
}