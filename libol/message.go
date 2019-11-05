package libol

import (
	"bytes"
	"fmt"
)

//[MAGIC(2)][Length(2)][DSTMAC(6)]
// if DSTMAC is ZERO
//    [Action(4+(=/:))[Space(1)][Json Body]
//    Cmd: Instruct such as 'logi=', 'logi:'.
//    Json Params: length - 6 bytes.
// else
//    Payload is Ethernet Frame.

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

func EncodeRequestCmd(action string, body string) []byte {
	p := fmt.Sprintf("%s= %s", action[:4], body)
	return append(ZEROED[:6], p...)
}

func EncodeReplyCmd(action string, body string) []byte {
	p := fmt.Sprintf("%s: %s", action[:4], body)
	return append(ZEROED[:6], p...)
}
