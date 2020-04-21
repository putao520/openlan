package ctl

import (
	"fmt"
	"golang.org/x/net/websocket"
	"strings"
)

type Message struct {
	Raw      string
	Action   string
	Resource string
	Data     string
}

func (m *Message) Encode() string {
	if m.Action == "" {
		m.Action = "GET"
	} else {
		m.Action = strings.ToUpper(m.Action)
	}
	m.Resource = strings.ToUpper(m.Resource)
	return fmt.Sprintf("%s %s %s", m.Action, m.Resource, m.Data)
}

func (m *Message) Decode() (string, string, string) {
	values := strings.SplitN(m.Raw, " ", 3)
	if len(values) == 3 {
		return values[0], values[1], values[2]
	} else if len(values) == 2 {
		return values[0], values[1], ""
	} else {
		return values[0], "", ""
	}
}

func (m *Message) String() string {
	if m.Raw == "" {
		return fmt.Sprintf("%s %s %s", m.Action, m.Resource, m.Data)
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
		data.Action, data.Resource, data.Data = m.Decode()
		return nil
	}
	return websocket.ErrNotSupported
}

var Codec = websocket.Codec{Marshal: marshal, Unmarshal: unmarshal}
