package libol

import (
	"bytes"
	"fmt"
)

//[MAGIC(2)][Length(2)][DSTMAC(6)]
// if DSTMAC is ZERO
//    [Action(4+(=/:))[Space(1)][Json Body]
//    Action: Instruct such as 'logi=', 'logi:'.
//    Json Body: length - 6 bytes.
// else
//    Payload is Ethernat Frame.

func IsInst(data []byte) bool {
	return bytes.Equal(data[:6], ZEROED[:6])
}

func DecAction(data []byte) string {
	return string(data[6:11])
}

func DecBody(data []byte) string {
	return string(data[12:])
}

func DecActionBody(data []byte) (string, string) {
	return DecAction(data), DecBody(data)
}

func EncInstReq(action string, body string) []byte {
	p := fmt.Sprintf("%s= %s", action[:4], body)
	return append(ZEROED[:6], p...)
}

func EncInstResp(action string, body string) []byte {
	p := fmt.Sprintf("%s: %s", action[:4], body)
	return append(ZEROED[:6], p...)
}
