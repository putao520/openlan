package libol

import (
	"bytes"
	"fmt"
)

const HSIZE = 0x04

func GetHeaderLen() int {
	return HSIZE
}

var MAGIC = []byte{0xff, 0xff}

func GetMagic() []byte {
	return MAGIC
}

func IsControl(data []byte) bool {
	if bytes.Equal(data[:6], ZEROED[:6]) {
		return true
	}
	return false
}

type FrameMessage struct {
	control bool
	action  string
	params  string
	frame   []byte
	raw     []byte
}

func NewFrameMessage(data []byte) *FrameMessage {
	m := FrameMessage{
		control: false,
		action:  "",
		params:  "",
		raw:     data,
	}
	m.Decode()
	return &m
}

func (m *FrameMessage) Decode() bool {
	m.control = IsControl(m.raw)
	if m.control {
		m.action = string(m.raw[6:11])
		m.params = string(m.raw[12:])
	} else {
		m.frame = m.raw
	}
	return m.control
}

func (m *FrameMessage) IsControl() bool {
	return m.control
}

func (m *FrameMessage) Data() []byte {
	return m.frame
}

func (m *FrameMessage) String() string {
	return fmt.Sprintf("control: %t,raw:% x", m.control, m.raw[:20])
}

func (m *FrameMessage) CmdAndParams() (string, string) {
	return m.action, m.params
}

type ControlMessage struct {
	control bool
	opr     string
	action  string
	params  string
}

//opr: request is '= ', and response is  ': '
//action: login, network etc.
//body: json string.
func NewControlMessage(action string, opr string, body string) *ControlMessage {
	c := ControlMessage{
		control: true,
		action:  action,
		params:  body,
		opr:     opr,
	}

	return &c
}

func (c *ControlMessage) Encode() []byte {
	p := fmt.Sprintf("%s%s%s", c.action[:4], c.opr[:2], c.params)
	return append(ZEROED[:6], p...)
}
