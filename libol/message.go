package libol

import (
	"bytes"
	"fmt"
)

const HSIZE = 0x04
var MAGIC = []byte{0xff, 0xff}

func IsControl(data []byte) bool {
	return bytes.Equal(data[:6], ZEROED[:6])
}

func DecodeCmd(data []byte) string {
	return string(data[6:11])
}

func DecodeParams(data []byte) string {
	return string(data[12:])
}

func DecodeCmdAndParams(data []byte) (string, string) {
	return DecodeCmd(data), DecodeParams(data)
}

func EncodeRequest(action string, body string) []byte {
	p := fmt.Sprintf("%s= %s", action[:4], body)
	return append(ZEROED[:6], p...)
}

func EncodeReply(action string, body string) []byte {
	p := fmt.Sprintf("%s: %s", action[:4], body)
	return append(ZEROED[:6], p...)
}
