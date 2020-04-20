package ctl

import (
	"fmt"
	"golang.org/x/net/websocket"
	"strings"
)

type Message struct {
	Raw  string
	Type string
	Data string
}

func (m *Message) Encode() string {
	return fmt.Sprintf("%s %s", m.Type, m.Data)
}

func (m *Message) Decode() (string, string) {
	if strings.Contains(m.Raw, " ") {
		values := strings.SplitN(m.Raw, " ", 2)
		return values[0], values[1]
	}
	return m.Raw, ""
}

func (m *Message) String() string {
	if m.Raw == "" {
		return fmt.Sprintf("%s: %s", m.Type, m.Data)
	}
	return m.Raw
}

func marshal(v interface{}) (msg []byte, payloadType byte, err error) {
	if data, ok := v.(*Message); ok {
		return []byte(data.Encode()), websocket.TextFrame, nil
	}
	return msg, websocket.UnknownFrame, websocket.ErrNotSupported
}

func unmarshal(msg []byte, payloadType byte, v interface{}) (err error) {
	if data, ok := v.(*Message); ok {
		m := &Message{
			Raw: string(msg),
		}
		data.Type, data.Data = m.Decode()
		return nil
	}
	return websocket.ErrNotSupported
}

var Codec = websocket.Codec{Marshal: marshal, Unmarshal: unmarshal}
