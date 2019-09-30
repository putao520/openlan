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
    return bytes.Equal(data[:6], ZEROETHADDR[:6])
}

func DecAction(data []byte) string {
    return string(data[6:11])
}

func DecBody(data []byte) string {
    return string(data[12:])
}

func EncInstReq(action string, body string) []byte {
    payload := fmt.Sprintf("%s= %s", action[:4], body)
    return append(ZEROETHADDR[:6], payload...)
}

func EncInstResp(action string, body string) []byte {
    payload := fmt.Sprintf("%s: %s", action[:4], body)
    return append(ZEROETHADDR[:6], payload...)
}