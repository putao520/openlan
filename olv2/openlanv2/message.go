package openlanv2

import (
	"bytes"
	"fmt"
)

var (
    MAGIC   = []byte{0xff,0xff}
	HSIZE   = uint16(0x14)
	CTLUUID = []byte{0x00, 0x00, 0x00, 0x00,
					 0x00, 0x00, 0x00, 0x00,
					 0x00, 0x00, 0x00, 0x00,
					 0x00, 0x00, 0x00, 0x00} 
)

type Message struct {
	Action string
	Body string
}

func NewMessage(action string, body string) (this *Message) {
	this = &Message {
		Action: action, 
		Body: body,
	}
	return
}

//[MAGIC(2)][Length(2)]
//[UUID(16)][DSTMAC(6)]
// if DSTMAC is ZERO
//    [Action(4+(=/:))[Space(1)][Json Body]
//    Action: Instruct such as 'logi=', 'logi:'.
//    Json Body: length - 6 bytes.
// else 
//    Payload is Ethernat Frame.

func IsInstruct(data []byte) bool {
	return bytes.Equal(data[:6], ZEROMAC)
}

func DecodeAction(data []byte) string {
	return string(data[6:11])
}

func DecodeBody(data []byte) string {
	return string(data[12:])
}

func (this *Message)EncodeReq() []byte {
	payload := fmt.Sprintf("%s= %s", this.Action[:4], this.Body)
	return append(ZEROMAC, payload...)
}

func (this *Message)EncodeResp() []byte {
	payload := fmt.Sprintf("%s: %s", this.Action[:4], this.Body)
	return append(ZEROMAC, payload...)
}